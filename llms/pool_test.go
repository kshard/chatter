//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llms_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	proxy "github.com/kshard/chatter/llms"
)

type mock struct {
	reply chatter.Reply
	err   error
}

func (m *mock) UsedInputTokens() int { return 0 }
func (m *mock) UsedReplyTokens() int { return 0 }

func (m *mock) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (chatter.Reply, error) {
	return m.reply, m.err
}

func TestProxy_Prompt(t *testing.T) {
	llms := map[string]chatter.Chatter{
		"model1": &mock{
			reply: chatter.Reply{
				Text:            "model1",
				UsedInputTokens: 10,
				UsedReplyTokens: 20,
			},
		},
	}
	fallback := &mock{
		reply: chatter.Reply{
			Text:            "fallback",
			UsedInputTokens: 5,
			UsedReplyTokens: 15,
		},
	}
	p := proxy.New(llms, fallback)

	t.Run("Routing", func(t *testing.T) {
		reply, err := p.Prompt(context.Background(), nil, proxy.Model("model1"))
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(reply.UsedInputTokens, 10),
			it.Equal(reply.UsedReplyTokens, 20),
			it.Equal(reply.Text, "model1"),
		)
	})

	t.Run("Unknown", func(t *testing.T) {
		reply, err := p.Prompt(context.Background(), nil, proxy.Model("unkown"))
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(reply.UsedInputTokens, 5),
			it.Equal(reply.UsedReplyTokens, 15),
			it.Equal(reply.Text, "fallback"),
		)
	})

	t.Run("Default", func(t *testing.T) {
		reply, err := p.Prompt(context.Background(), nil)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(reply.UsedInputTokens, 5),
			it.Equal(reply.UsedReplyTokens, 15),
			it.Equal(reply.Text, "fallback"),
		)
	})

}
