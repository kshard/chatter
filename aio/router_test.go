//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package aio_test

import (
	"context"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio"
)

func TestProxy_Prompt(t *testing.T) {
	llms := map[string]chatter.Chatter{
		"model1": &mock{&chatter.Reply{
			Content: []chatter.Content{chatter.Text("model1")},
			Usage:   chatter.Usage{InputTokens: 10, ReplyTokens: 20},
		}},
	}
	fallback := &mock{&chatter.Reply{
		Content: []chatter.Content{chatter.Text("fallback")},
		Usage:   chatter.Usage{InputTokens: 5, ReplyTokens: 15},
	}}
	p := aio.NewRouter(llms, fallback)

	t.Run("Routing", func(t *testing.T) {
		reply, err := p.Prompt(context.Background(), nil, aio.Route("model1"))
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(reply.Usage.InputTokens, 10),
			it.Equal(reply.Usage.ReplyTokens, 20),
			it.Equal(reply.String(), "model1"),
		)
	})

	t.Run("Unknown", func(t *testing.T) {
		reply, err := p.Prompt(context.Background(), nil, aio.Route("unkown"))
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(reply.Usage.InputTokens, 5),
			it.Equal(reply.Usage.ReplyTokens, 15),
			it.Equal(reply.String(), "fallback"),
		)
	})

	t.Run("Default", func(t *testing.T) {
		reply, err := p.Prompt(context.Background(), nil)
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(reply.Usage.InputTokens, 5),
			it.Equal(reply.Usage.ReplyTokens, 15),
			it.Equal(reply.String(), "fallback"),
		)
	})
}
