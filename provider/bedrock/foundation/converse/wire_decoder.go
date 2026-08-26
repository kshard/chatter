//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import "github.com/kshard/chatter"

type wireDecoder struct{}

func (wireDecoder) Decode(out *wireOutput) (*chatter.Reply, error) {
	reply := new(chatter.Reply)
	reply.Stage = decodeWireStage(out.StopReason)

	if out.Output != nil {
		content := make([]chatter.Content, 0, len(out.Output.Message.Content))
		for _, block := range out.Output.Message.Content {
			if c := decodeWireContent(block); c != nil {
				content = append(content, c)
			}
		}
		reply.Content = content
	}

	reply.Usage.InputTokens = out.Usage.InputTokens
	reply.Usage.ReplyTokens = out.Usage.OutputTokens

	return reply, nil
}

func decodeWireStage(reason string) chatter.Stage {
	switch reason {
	case "end_turn":
		return chatter.LLM_RETURN
	case "max_tokens", "stop_sequence":
		return chatter.LLM_INCOMPLETE
	case "tool_use":
		return chatter.LLM_INVOKE
	default:
		return chatter.LLM_ERROR
	}
}

func decodeWireContent(block wireContent) chatter.Content {
	switch {
	case block.ToolUse != nil:
		return chatter.Invoke{
			Cmd:  block.ToolUse.Name,
			Args: chatter.Json{ID: block.ToolUse.ToolUseID, Value: block.ToolUse.Input},
		}
	case block.Text != "":
		return chatter.Text(block.Text)
	default:
		return nil
	}
}
