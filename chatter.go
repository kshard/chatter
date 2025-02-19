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

type Opt = interface{ ChatterOpt() }

// The generic trait to "interact" with LLMs;
type Chatter interface {
	UsedInputTokens() int
	UsedReplyTokens() int
	Prompt(context.Context, []fmt.Stringer, ...Opt) (Reply, error)
}

// The reply from LLMs
type Reply struct {
	Text            string
	UsedInputTokens int
	UsedReplyTokens int
}

func (reply Reply) String() string { return reply.Text }

// LLMs' critical parameter influencing the balance between predictability
// and creativity in generated text. Lower temperatures prioritize exploiting
// learned patterns, yielding more deterministic outputs, while higher
// temperatures encourage exploration, fostering diversity and innovation.
type Temperature float64

func (Temperature) ChatterOpt() {}

// Nucleus Sampling, a parameter used in LLMs, impacts token selection by
// considering only the most likely tokens that together represent
// a cumulative probability mass (e.g., top-p tokens). This limits the
// number of choices to avoid overly diverse or nonsensical outputs while
// maintaining diversity within the top-ranked options.
type TopP float64

func (TopP) ChatterOpt() {}

// Token quota for reply, the model would limit response given number
type Quota int

func (Quota) ChatterOpt() {}

// The stop sequence prevents LLMsfrom generating more text after a specific
// string appears. Stop sequences make it easy to guarantee concise,
// controlled responses from models.
type StopSequence string

func (StopSequence) ChatterOpt() {}

// Foundational identity of LLMs
type LLM interface {
	// Model ID as defined by the vendor
	ModelID() string

	// Encode prompt to bytes:
	// - encoding prompt as prompt markup supported by LLM
	// - encoding prompt to envelop supported by LLM's hosting platform
	Encode([]fmt.Stringer, ...Opt) ([]byte, error)

	// Decode LLM's reply into pure text
	Decode([]byte) (Reply, error)
}

//------------------------------------------------------------------------------

// Debugger LLM's I/O
type Debugger struct {
	Chatter
	w io.Writer
}

// Create new debugger session
func NewDebugger(w io.Writer, chatter Chatter) *Debugger {
	return &Debugger{
		Chatter: chatter,
		w:       w,
	}
}

func (deb *Debugger) Prompt(ctx context.Context, seq []fmt.Stringer, opt ...Opt) (Reply, error) {
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
		deb.w.Write([]byte(reply.Text))
		deb.w.Write([]byte("\n"))
	}

	return reply, err
}
