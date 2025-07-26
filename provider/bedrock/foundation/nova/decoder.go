//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package nova

import (
	"github.com/kshard/chatter"
)

func (decoder decoder) Decode(response *reply) (*chatter.Reply, error) {
	reply := new(chatter.Reply)

	switch response.StopReason {
	case "end_turn":
		reply.Stage = chatter.LLM_RETURN
	case "max_tokens":
		reply.Stage = chatter.LLM_INCOMPLETE
	case "stop_sequence":
		reply.Stage = chatter.LLM_INCOMPLETE
	default:
		reply.Stage = chatter.LLM_ERROR
	}

	reply.Usage.InputTokens = response.Usage.InputTokens
	reply.Usage.ReplyTokens = response.Usage.OutputTokens

	reply.Content = make([]chatter.Content, 0, len(response.Output.Message.Content))
	for _, c := range response.Output.Message.Content {
		if c.Text != "" {
			reply.Content = append(reply.Content, chatter.Text(c.Text))
		}
		// TODO: Image and Video content handling
	}

	return reply, nil
}
