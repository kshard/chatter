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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kshard/chatter"
)

// Mock is a simple mock LLM that echoes the input.
type Mock struct {
	usage chatter.Usage
	reply any
}

func NewMock(reply any) *Mock {
	return &Mock{reply: reply}
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
	var reply string
	switch r := m.reply.(type) {
	case nil:
		reply = strings.Join(seq, " ")
	case string:
		reply = r
	case fmt.Stringer:
		reply = r.String()
	default:
		b, err := json.Marshal(m.reply)
		if err != nil {
			return nil, err
		}
		reply = string(b)
	}

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
