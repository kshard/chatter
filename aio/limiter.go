//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package aio

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/kshard/chatter"
	"golang.org/x/time/rate"
)

// Rate limit startegy for LLMs I/O
type Limiter struct {
	chatter.Chatter
	debt int
	rps  *rate.Limiter
	tps  *rate.Limiter
}

var _ chatter.Chatter = (*Limiter)(nil)

// Create rate limit strategy for LLMs.
// It defines per minute policy for requests and tokens.
func NewLimiter(requestPerMin int, tokensPerMin int, chatter chatter.Chatter) *Limiter {
	return &Limiter{
		Chatter: chatter,
		debt:    0,
		rps:     rate.NewLimiter(rate.Limit(requestPerMin)/60, requestPerMin),
		tps:     rate.NewLimiter(rate.Limit(tokensPerMin)/60, tokensPerMin),
	}
}

func (c *Limiter) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (chatter.Reply, error) {
	if err := c.rps.Wait(ctx); err != nil {
		return chatter.Reply{}, err
	}

	if err := c.tps.WaitN(ctx, c.debt); err != nil {
		return chatter.Reply{}, err
	}

	reply, err := c.Chatter.Prompt(ctx, prompt, opts...)
	if err != nil {
		return chatter.Reply{}, err
	}

	c.debt = reply.UsedInputTokens + reply.UsedReplyTokens

	slog.Debug("LLM is prompted",
		slog.Float64("budget", c.tps.Tokens()),
		slog.Int("debt", c.debt),
		slog.Group("session",
			slog.Int("inputTokens", c.Chatter.UsedInputTokens()),
			slog.Int("replyTokens", c.Chatter.UsedReplyTokens()),
		),
		slog.Group("prompt",
			slog.Int("inputTokens", reply.UsedInputTokens),
			slog.Int("replyTokens", reply.UsedReplyTokens),
		),
	)

	return reply, nil
}
