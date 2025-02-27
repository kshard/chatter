//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llms

import (
	"context"
	"fmt"

	"github.com/kshard/chatter"
)

// LLM pool specific parameter allowing to dynamically route request to models
type Model string

func (Model) ChatterOpt() {}

// LLM pool consists of default "route" and multiple named models.
type Pool struct {
	llms            map[string]chatter.Chatter
	fallback        chatter.Chatter
	usedInputTokens int
	usedReplyTokens int
}

// Creates LLMs pools instance
func New(llms map[string]chatter.Chatter, fallback chatter.Chatter) *Pool {
	return &Pool{
		llms:     llms,
		fallback: fallback,
	}
}

func (p *Pool) UsedInputTokens() int { return p.usedInputTokens }
func (p *Pool) UsedReplyTokens() int { return p.usedReplyTokens }

func (p *Pool) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (chatter.Reply, error) {
	llm := p.fallback

	for _, opt := range opts {
		switch v := opt.(type) {
		case Model:
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
