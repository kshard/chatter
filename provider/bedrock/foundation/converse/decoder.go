//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import (
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/kshard/chatter"
)

func (decoder decoder) Decode(result *bedrockruntime.ConverseOutput) (*chatter.Reply, error) {
	reply := new(chatter.Reply)
	reply.Stage = decodeStage(result.StopReason)

	content, err := decodeOutput(result.Output)
	if err != nil {
		return nil, err
	}
	reply.Content = content

	if result.Usage != nil {
		reply.Usage.InputTokens += int(aws.ToInt32(result.Usage.InputTokens))
		reply.Usage.ReplyTokens += int(aws.ToInt32(result.Usage.OutputTokens))
	}

	return reply, nil
}

func decodeStage(reason types.StopReason) chatter.Stage {
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

func decodeOutput(out types.ConverseOutput) ([]chatter.Content, error) {
	switch v := out.(type) {
	case *types.ConverseOutputMemberMessage:
		seq := make([]chatter.Content, 0)
		for _, block := range v.Value.Content {
			c, err := decodeContent(block)
			if err != nil {
				return nil, err
			}
			if c != nil {
				seq = append(seq, c)
			}
		}
		return seq, nil
	default:
		return nil, fmt.Errorf("unknown bedrock api response %T", out)
	}
}

func decodeContent(block types.ContentBlock) (chatter.Content, error) {
	switch v := block.(type) {
	case *types.ContentBlockMemberText:
		return chatter.Text(v.Value), nil

	case *types.ContentBlockMemberToolUse:
		in, err := v.Value.Input.MarshalSmithyDocument()
		if err != nil {
			return nil, err
		}
		return chatter.Invoke{
			Cmd: aws.ToString(v.Value.Name),
			Args: chatter.Json{
				ID:    aws.ToString(v.Value.ToolUseId),
				Value: in,
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
