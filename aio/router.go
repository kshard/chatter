//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package aio

import (
	"context"
	"fmt"

	"github.com/kshard/chatter"
)

// Chatter interface option allowing to dynamically route prompts to choosen models
type Route string

func (Route) ChatterOpt() {}

// Dynamic routing strategy throught pool of LLMs.
// The LLMs pool consists of default "route" and multiple named models.
type Router struct {
	llms            map[string]chatter.Chatter
	fallback        chatter.Chatter
	usedInputTokens int
	usedReplyTokens int
}

var _ chatter.Chatter = (*Router)(nil)

// Creates LLMs pools instance
func NewRouter(llms map[string]chatter.Chatter, fallback chatter.Chatter) *Router {
	return &Router{
		llms:     llms,
		fallback: fallback,
	}
}

func (p *Router) UsedInputTokens() int { return p.usedInputTokens }
func (p *Router) UsedReplyTokens() int { return p.usedReplyTokens }

func (p *Router) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (chatter.Reply, error) {
	llm := p.fallback

	for _, opt := range opts {
		switch v := opt.(type) {
		case Route:
			if l, has := p.llms[string(v)]; has {
				llm = l
			}
		}
	}

	reply, err := llm.Prompt(ctx, prompt, opts...)
	if err != nil {
		return reply, err
	}

	p.usedInputTokens += reply.UsedInputTokens
	p.usedReplyTokens += reply.UsedReplyTokens

	return reply, nil
}
