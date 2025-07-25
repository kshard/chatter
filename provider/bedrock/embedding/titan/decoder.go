//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package titan

import "github.com/kshard/chatter"

func (decoder decoder) Decode(titan *reply) (*chatter.Reply, error) {
	reply := new(chatter.Reply)
	reply.Stage = chatter.LLM_RETURN

	reply.Usage.InputTokens = titan.UsedTextTokens

	reply.Content = []chatter.Content{
		chatter.Vector(titan.Vector),
	}

	return reply, nil
}
