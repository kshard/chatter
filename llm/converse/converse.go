//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/fogfish/opts"
	"github.com/kshard/chatter"
)

type LLM string

type Bedrock interface {
	Converse(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error)
}

// Command registry
type Registry []chatter.Cmd

func (Registry) ChatterOpt() {}

type Client struct {
	api             Bedrock
	llm             LLM
	registry        Registry
	usedInputTokens int
	usedReplyTokens int
}

func New(opt ...Option) (*Client, error) {
	c := Client{}

	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}

	if c.api == nil {
		if err := optsFromRegion(&c, defaultRegion); err != nil {
			return nil, err
		}
	}

	return &c, c.checkRequired()

}

func (c *Client) UsedInputTokens() int { return c.usedInputTokens }
func (c *Client) UsedReplyTokens() int { return c.usedReplyTokens }

func (c *Client) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (reply *chatter.Reply, err error) {
	if len(prompt) == 0 {
		err = fmt.Errorf("bad request, empty prompt")
		return
	}

	inquiry := &bedrockruntime.ConverseInput{
		InferenceConfig: &types.InferenceConfiguration{},
		ModelId:         (*string)(&c.llm),
		Messages:        make([]types.Message, 0),
		System:          make([]types.SystemContentBlock, 0),
	}

	if len(c.registry) > 0 {
		reg, err := toToolConfig(c.registry)
		if err != nil {
			return nil, err
		}
		inquiry.ToolConfig = reg
	}

	for _, opt := range opts {
		switch v := opt.(type) {
		case chatter.Temperature:
			inquiry.InferenceConfig.Temperature = aws.Float32(float32(v))
		case chatter.TopP:
			inquiry.InferenceConfig.TopP = aws.Float32(float32(v))
		case chatter.Quota:
			inquiry.InferenceConfig.MaxTokens = aws.Int32(int32(v))
		case Registry:
			reg, err := toToolConfig(v)
			if err != nil {
				return nil, err
			}
			inquiry.ToolConfig = reg
		}
	}

	for _, term := range prompt {
		switch v := term.(type) {
		case chatter.Stratum:
			inquiry.System = append(inquiry.System,
				&types.SystemContentBlockMemberText{Value: string(v)},
			)
		default:
			msg, err := toMessage(term)
			if err != nil {
				return nil, err
			}
			if len(msg.Content) != 0 {
				inquiry.Messages = append(inquiry.Messages, msg)
			}
		}
	}

	result, err := c.api.Converse(ctx, inquiry)
	if err != nil {
		return nil, err
	}

	r, err := toReply(result.Output)
	if err != nil {
		return nil, err
	}
	r.Stage = toStage(result.StopReason)

	if result.Usage != nil {
		if result.Usage.InputTokens != nil {
			r.Usage.InputTokens = int(*result.Usage.InputTokens)
			c.usedInputTokens += int(*result.Usage.InputTokens)
		}
		if result.Usage.OutputTokens != nil {
			r.Usage.ReplyTokens = int(*result.Usage.OutputTokens)
			c.usedReplyTokens += int(*result.Usage.OutputTokens)
		}
	}

	return &r, nil

}

func toStage(reason types.StopReason) chatter.Stage {
	switch reason {
	case types.StopReasonEndTurn:
		return chatter.LLM_RETURN
	case types.StopReasonMaxTokens:
		return chatter.LLM_INCOMPLETE
	case types.StopReasonStopSequence:
		return chatter.LLM_INCOMPLETE
	case types.StopReasonToolUse:
		return chatter.LLM_INVOKE
	default:
		return chatter.LLM_ERROR
	}
}

func toContent(block types.ContentBlock) (chatter.Content, error) {
	switch v := block.(type) {
	case *types.ContentBlockMemberText:
		return chatter.ContentText{Text: v.Value}, nil

	case *types.ContentBlockMemberToolUse:
		in, err := v.Value.Input.MarshalSmithyDocument()
		if err != nil {
			return nil, err
		}
		return chatter.Invoke{
			Name: aws.ToString(v.Value.Name),
			Args: chatter.ContentJson{
				Source: aws.ToString(v.Value.ToolUseId),
				Value:  in,
			},
			Message: v,
		}, nil
	default:
		slog.Warn("chatter does not support aws bedrock content type",
			slog.String("type", fmt.Sprintf("%T", block)),
		)
		return nil, nil
	}
}

func toReply(out types.ConverseOutput) (reply chatter.Reply, err error) {
	switch v := out.(type) {
	case *types.ConverseOutputMemberMessage:
		reply = chatter.Reply{Content: make([]chatter.Content, 0)}
		for _, block := range v.Value.Content {
			c, exx := toContent(block)
			if exx != nil {
				err = exx
				return
			}
			reply.Content = append(reply.Content, c)
		}
	default:
		err = fmt.Errorf("unknown bedrock api response %T", out)
	}

	return
}

func toMessage(msg fmt.Stringer) (types.Message, error) {
	switch v := (msg).(type) {
	case *chatter.Reply:
		msg := types.Message{
			Role:    types.ConversationRoleAssistant,
			Content: []types.ContentBlock{},
		}
		for _, block := range v.Content {
			switch b := (block).(type) {
			case chatter.ContentText:
				msg.Content = append(msg.Content, &types.ContentBlockMemberText{Value: b.Text})
			case interface{ RawMessage() any }:
				if cb, ok := b.RawMessage().(types.ContentBlock); ok {
					msg.Content = append(msg.Content, cb)
				}
			}
		}
		return msg, nil
	case chatter.Answer:
		msg := types.Message{
			Role:    types.ConversationRoleUser,
			Content: []types.ContentBlock{},
		}
		for _, yield := range v.Yield {
			var reply any
			if err := json.Unmarshal(yield.Value, &reply); err != nil {
				return types.Message{}, err
			}
			msg.Content = append(msg.Content, &types.ContentBlockMemberToolResult{
				Value: types.ToolResultBlock{
					ToolUseId: aws.String(yield.Source),
					Content: []types.ToolResultContentBlock{
						&types.ToolResultContentBlockMemberJson{
							Value: document.NewLazyDocument(
								map[string]any{
									"json": reply,
								},
							),
						},
					},
				},
			})
		}
		return msg, nil

	case *chatter.Prompt:
		return types.Message{
			Role: types.ConversationRoleUser,
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{Value: v.String()},
			},
		}, nil

	case chatter.Prompt:
		return types.Message{
			Role: types.ConversationRoleUser,
			Content: []types.ContentBlock{
				&types.ContentBlockMemberText{Value: v.String()},
			},
		}, nil
	}

	return types.Message{}, fmt.Errorf("invalid content block")
}

func toToolConfig(registry Registry) (*types.ToolConfiguration, error) {
	tools := &types.ToolConfiguration{
		ToolChoice: &types.ToolChoiceMemberAuto{},
		Tools:      []types.Tool{},
	}

	for _, reg := range registry {
		var spec any
		if err := json.Unmarshal(reg.Schema, &spec); err != nil {
			return nil, err
		}

		tools.Tools = append(tools.Tools,
			&types.ToolMemberToolSpec{
				Value: types.ToolSpecification{
					Name:        aws.String(reg.Cmd),
					Description: aws.String(reg.About),
					InputSchema: &types.ToolInputSchemaMemberJson{
						Value: document.NewLazyDocument(spec),
					},
				},
			},
		)
	}

	return tools, nil
}
