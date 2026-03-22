//
// Copyright (C) 2024 - 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package autoconfig

import (
	"context"
	"strings"

	"github.com/kshard/chatter"
)

// Mock is a simple mock LLM that echoes the input.
type Mock struct {
	usage chatter.Usage
}

func (m *Mock) Usage() chatter.Usage {
	return m.usage
}

func (m *Mock) Prompt(ctx context.Context, prompt []chatter.Message, opt ...chatter.Opt) (*chatter.Reply, error) {
	// Echo all messages
	seq := make([]string, len(prompt))
	for i, msg := range prompt {
		seq[i] = msg.String()
	}
	reply := strings.Join(seq, " ")

	m.usage.InputTokens += len(reply)
	m.usage.ReplyTokens += len(reply)

	return &chatter.Reply{
		Stage: chatter.LLM_RETURN,
		Usage: chatter.Usage{
			InputTokens: len(reply),
			ReplyTokens: len(reply),
		},
		Content: []chatter.Content{
			chatter.Text(reply),
		},
	}, nil
}
