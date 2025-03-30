//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package aio_test

import (
	"context"
	"fmt"

	"github.com/kshard/chatter"
)

type mock struct {
	reply chatter.Reply
}

func (mock mock) UsedInputTokens() int { return mock.reply.UsedInputTokens }
func (mock mock) UsedReplyTokens() int { return mock.reply.UsedReplyTokens }

func (mock mock) Prompt(context.Context, []fmt.Stringer, ...chatter.Opt) (chatter.Reply, error) {
	return mock.reply, nil
}

// // mock LLMs api
// type mock struct {
// 	reply  chatter.Reply
// 	err    error
// 	tokens int
// }

// func mockWithTokens(t int) *mock {
// 	return &mock{tokens: t}
// }

// func (mock mock) UsedInputTokens() int { return 0 }
// func (mock mock) UsedReplyTokens() int { return mock.tokens }

// func (mock mock) Prompt(context.Context, []fmt.Stringer, ...chatter.Opt) (chatter.Reply, error) {
// 	return chatter.Reply{
// 		Text:            "Looking for testing",
// 		UsedInputTokens: 0,
// 		UsedReplyTokens: m.int,
// 	}, nil
// }

/*

type mock struct {
	reply chatter.Reply
	err   error
}

func (m *mock) UsedInputTokens() int { return 0 }
func (m *mock) UsedReplyTokens() int { return 0 }

func (m *mock) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (chatter.Reply, error) {
	return m.reply, m.err
}


*/
