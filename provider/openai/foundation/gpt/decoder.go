//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package gpt

import "github.com/kshard/chatter"

func (decoder decoder) Decode(bag *reply) (*chatter.Reply, error) {
	reply := &chatter.Reply{
		Stage: chatter.LLM_RETURN,
		Content: []chatter.Content{
			chatter.Text(bag.Choices[0].Message.Content),
		},
		Usage: chatter.Usage{
			InputTokens: bag.Usage.PromptTokens,
			ReplyTokens: bag.Usage.OutputTokens,
		},
	}
	return reply, nil
}
