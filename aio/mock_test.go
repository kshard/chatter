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

	"github.com/kshard/chatter"
)

type mock struct {
	reply *chatter.Reply
}

func (mock mock) Usage() chatter.Usage { return mock.reply.Usage }

func (mock mock) Prompt(context.Context, []chatter.Message, ...chatter.Opt) (*chatter.Reply, error) {
	return mock.reply, nil
}
