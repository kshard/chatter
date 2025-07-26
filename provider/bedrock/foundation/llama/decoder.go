//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llama

import "github.com/kshard/chatter"

func (decoder decoder) Decode(bag *reply) (*chatter.Reply, error) {
	reply := new(chatter.Reply)

	switch bag.StopReason {
	case "stop":
		reply.Stage = chatter.LLM_RETURN
	case "length":
		reply.Stage = chatter.LLM_INCOMPLETE
	default:
		reply.Stage = chatter.LLM_ERROR
	}

	reply.Usage.InputTokens = bag.UsedPromptTokens
	reply.Usage.ReplyTokens = bag.UsedTextTokens

	reply.Content = []chatter.Content{
		chatter.Text(bag.Text),
	}

	return reply, nil
}
