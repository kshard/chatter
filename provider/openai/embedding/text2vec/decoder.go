//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text2vec

import (
	"errors"

	"github.com/kshard/chatter"
)

func (decoder decoder) Decode(bag *reply) (*chatter.Reply, error) {
	if len(bag.Vectors) != 1 {
		return nil, errors.New("invalid response")
	}

	reply := new(chatter.Reply)
	reply.Stage = chatter.LLM_RETURN

	reply.Usage.InputTokens = bag.Usage.UsedTokens

	reply.Content = []chatter.Content{
		chatter.Vector(bag.Vectors[0].Vector),
	}

	return reply, nil
}
