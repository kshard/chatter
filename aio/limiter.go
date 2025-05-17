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

func (c *Limiter) Prompt(ctx context.Context, prompt []chatter.Message, opts ...chatter.Opt) (*chatter.Reply, error) {
	if err := c.rps.Wait(ctx); err != nil {
		return nil, err
	}

	if err := c.tps.WaitN(ctx, c.debt); err != nil {
		return nil, err
	}

	reply, err := c.Chatter.Prompt(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}

	c.debt = reply.Usage.InputTokens + reply.Usage.ReplyTokens

	slog.Debug("LLM is prompted",
		slog.Float64("budget", c.tps.Tokens()),
		slog.Int("debt", c.debt),
		slog.Group("session",
			slog.Int("inputTokens", c.Chatter.Usage().InputTokens),
			slog.Int("replyTokens", c.Chatter.Usage().ReplyTokens),
		),
		slog.Group("prompt",
			slog.Int("inputTokens", reply.Usage.InputTokens),
			slog.Int("replyTokens", reply.Usage.ReplyTokens),
		),
	)

	return reply, nil
}
