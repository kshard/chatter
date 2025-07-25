package converse

import (
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory(model string, registry chatter.Registry) func() (provider.Encoder[*bedrockruntime.ConverseInput], error) {
	return func() (provider.Encoder[*bedrockruntime.ConverseInput], error) {
		var toolConfig *types.ToolConfiguration
		if len(registry) > 0 {
			reg, err := encodeRegistry(registry)
			if err != nil {
				return nil, err
			}
			toolConfig = reg
		}

		return &encoder{
			req: &bedrockruntime.ConverseInput{
				InferenceConfig: &types.InferenceConfiguration{},
				ModelId:         aws.String(model),
				Messages:        make([]types.Message, 0),
				System:          make([]types.SystemContentBlock, 0),
				ToolConfig:      toolConfig,
			},
		}, nil
	}
}

func (codec *encoder) WithInferrer(inferrer provider.Inferrer) {
	codec.req.InferenceConfig.Temperature = aws.Float32(float32(inferrer.Temperature))
	codec.req.InferenceConfig.TopP = aws.Float32(float32(inferrer.TopP))
	codec.req.InferenceConfig.MaxTokens = aws.Int32(int32(inferrer.MaxTokens))
	codec.req.InferenceConfig.StopSequences = inferrer.StopSequences
}

func (codec *encoder) WithCommand(cmd chatter.Cmd) {
	tool, err := encodeCommand(cmd)
	if err != nil {
		slog.Warn("invalid tool configuration", "cmd", cmd.Cmd, "err", err)
		return
	}
	if codec.req.ToolConfig == nil {
		codec.req.ToolConfig = &types.ToolConfiguration{
			ToolChoice: &types.ToolChoiceMemberAuto{},
			Tools:      []types.Tool{},
		}
	}

	codec.req.ToolConfig.Tools = append(codec.req.ToolConfig.Tools, tool)
}

// AsStratum processes a Stratum message (system role)
func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	codec.req.System = append(codec.req.System,
		&types.SystemContentBlockMemberText{Value: string(stratum)},
	)
	return nil
}

// AsText processes a Text message as user input
func (codec *encoder) AsText(text chatter.Text) error {
	msg := types.Message{
		Role: types.ConversationRoleUser,
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{Value: string(text)},
		},
	}

	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

// AsPrompt processes a Prompt message by converting it to string
func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	msg := types.Message{
		Role: types.ConversationRoleUser,
		Content: []types.ContentBlock{
			&types.ContentBlockMemberText{Value: prompt.String()},
		},
	}

	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

// AsAnswer processes an Answer message (tool results)
func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	if len(answer.Yield) == 0 {
		return nil
	}

	msg := types.Message{
		Role:    types.ConversationRoleUser,
		Content: []types.ContentBlock{},
	}
	for _, yield := range answer.Yield {
		var reply any
		if err := json.Unmarshal(yield.Value, &reply); err != nil {
			return err
		}
		msg.Content = append(msg.Content,
			&types.ContentBlockMemberToolResult{
				Value: types.ToolResultBlock{
					ToolUseId: aws.String(yield.ID),
					Content: []types.ToolResultContentBlock{
						&types.ToolResultContentBlockMemberJson{
							Value: document.NewLazyDocument(
								map[string]any{"json": reply},
							),
						},
					},
				},
			},
		)
	}

	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

// AsReply processes a Reply message (assistant response)
func (codec *encoder) AsReply(reply *chatter.Reply) error {
	msg := types.Message{
		Role:    types.ConversationRoleAssistant,
		Content: []types.ContentBlock{},
	}

	for _, block := range reply.Content {
		switch v := (block).(type) {
		case chatter.Text:
			msg.Content = append(msg.Content,
				&types.ContentBlockMemberText{Value: string(v)},
			)
		case interface{ RawMessage() any }:
			if cb, ok := v.RawMessage().(types.ContentBlock); ok {
				msg.Content = append(msg.Content, cb)
			}
		}
	}

	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

func (codec *encoder) Build() *bedrockruntime.ConverseInput {
	return codec.req
}

func encodeRegistry(registry chatter.Registry) (*types.ToolConfiguration, error) {
	tools := &types.ToolConfiguration{
		ToolChoice: &types.ToolChoiceMemberAuto{},
		Tools:      []types.Tool{},
	}

	for _, reg := range registry {
		cmd, err := encodeCommand(reg)
		if err != nil {
			return nil, err
		}
		tools.Tools = append(tools.Tools, cmd)
	}

	return tools, nil
}

func encodeCommand(cmd chatter.Cmd) (types.Tool, error) {
	var spec any
	if err := json.Unmarshal(cmd.Schema, &spec); err != nil {
		return nil, err
	}

	return &types.ToolMemberToolSpec{
		Value: types.ToolSpecification{
			Name:        aws.String(cmd.Cmd),
			Description: aws.String(cmd.About),
			InputSchema: &types.ToolInputSchemaMemberJson{
				Value: document.NewLazyDocument(spec),
			},
		},
	}, nil
}
