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
