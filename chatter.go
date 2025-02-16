//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

import (
	"context"
	"fmt"
	"io"
)

// The generic trait to "interact" with LLMs;
type Chatter interface {
	UsedInputTokens() int
	UsedReplyTokens() int
	Prompt(context.Context, []fmt.Stringer, ...func(*Options)) (Text, error)
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

	// Token quota for reply, the model would limit response given number
	Quota int
}

// Return default options chatter
func NewOptions() Options {
	return Options{
		Temperature: 0.5,
		TopP:        0.9,
	}
}

// LLMs' critical parameter influencing the balance between predictability
// and creativity in generated text. Lower temperatures prioritize exploiting
// learned patterns, yielding more deterministic outputs, while higher
// temperatures encourage exploration, fostering diversity and innovation.
func WithTemperature(t float64) func(*Options) {
	return func(opt *Options) { opt.Temperature = t }
}

// Nucleus Sampling, a parameter used in LLMs, impacts token selection by
// considering only the most likely tokens that together represent
// a cumulative probability mass (e.g., top-p tokens). This limits the
// number of choices to avoid overly diverse or nonsensical outputs while
// maintaining diversity within the top-ranked options.
func WithTopP(p float64) func(*Options) {
	return func(opt *Options) { opt.TopP = p }
}

// Limit reply to given quota
func WithQuota(quota int) func(*Options) {
	return func(opt *Options) { opt.Quota = quota }
}

// Foundational identity of LLMs
type LLM interface {
	// Model ID as defined by the vendor
	ModelID() string

	// Encode prompt to bytes:
	// - encoding prompt as prompt markup supported by LLM
	// - encoding prompt to envelop supported by LLM's hosting platform
	Encode([]fmt.Stringer, *Options) ([]byte, error)

	// Decode LLM's reply into pure text
	Decode([]byte) (Reply, error)
}

type Reply struct {
	Text            Text
	UsedInputTokens int
	UsedReplyTokens int
}

type Text string

func (txt Text) String() string { return string(txt) }

//------------------------------------------------------------------------------

type Debugger struct {
	Chatter
	w io.Writer
}

func NewDebugger(w io.Writer, chatter Chatter) *Debugger {
	return &Debugger{
		Chatter: chatter,
		w:       w,
	}
}

func (deb *Debugger) Prompt(ctx context.Context, seq []fmt.Stringer, opt ...func(*Options)) (Text, error) {
	if len(seq) != 0 {
		ask := seq[len(seq)-1].String()
		deb.w.Write([]byte("\n>>>>>\n"))
		deb.w.Write([]byte(ask))
		deb.w.Write([]byte("\n"))
	}

	reply, err := deb.Chatter.Prompt(ctx, seq, opt...)
	if err != nil {
		deb.w.Write([]byte("FAIL:\n\t"))
		deb.w.Write([]byte(err.Error()))
		deb.w.Write([]byte("\n"))
	} else {
		deb.w.Write([]byte("\n<<<<<\n"))
		deb.w.Write([]byte(reply))
		deb.w.Write([]byte("\n"))
	}

	return reply, err
}
