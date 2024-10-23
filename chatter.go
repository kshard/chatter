//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

import "context"

// The generic trait to "interact" with LLMs;
type Chatter interface {
	ConsumedTokens() int
	Prompt(context.Context, *Prompt, ...func(*Options)) (string, error)
}

type Options struct {
	// LLMs' critical parameter influencing the balance between predictability
	// and creativity in generated text. Lower temperatures prioritize exploiting
	// learned patterns, yielding more deterministic outputs, while higher
	// temperatures encourage exploration, fostering diversity and innovation.
	Temperature float64

	// Nucleus Sampling, a parameter used in LLMs, impacts token selection by
	// considering only the most likely tokens that together represent
	// a cumulative probability mass (e.g., top-p tokens). This limits the
	// number of choices to avoid overly diverse or nonsensical outputs while
	// maintaining diversity within the top-ranked options.
	TopP float64
}

const DefaultTemperature = 0.5

// LLMs' critical parameter influencing the balance between predictability
// and creativity in generated text. Lower temperatures prioritize exploiting
// learned patterns, yielding more deterministic outputs, while higher
// temperatures encourage exploration, fostering diversity and innovation.
func WithTemperature(t float64) func(*Options) {
	return func(opt *Options) {
		opt.Temperature = t
	}
}

const DefaultTopP = 0.9

// Nucleus Sampling, a parameter used in LLMs, impacts token selection by
// considering only the most likely tokens that together represent
// a cumulative probability mass (e.g., top-p tokens). This limits the
// number of choices to avoid overly diverse or nonsensical outputs while
// maintaining diversity within the top-ranked options.
func WithTopP(p float64) func(*Options) {
	return func(opt *Options) {
		opt.TopP = p
	}
}
