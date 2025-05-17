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
	"encoding/json"
)

type Opt = interface{ ChatterOpt() }

// The generic trait to "interact" with LLMs;
type Chatter interface {
	Usage() Usage
	Prompt(context.Context, []Message, ...Opt) (*Reply, error)
}

// LLM Usage stats
type Usage struct {
	InputTokens int `json:"inputTokens"`
	ReplyTokens int `json:"replyTokens"`
}

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

// Command registry is a sequence of tools available for LLM usage.
type Registry []Cmd

func (Registry) ChatterOpt() {}

// Command descriptor
type Cmd struct {
	// [Required] A unique name for the command, used as a reference by LLMs (e.g., "bash").
	Cmd string `json:"cmd"`

	// [Required] A detailed, multi-line description to educate the LLM on command usage.
	// Provides contextual information on how and when to use the command.
	About string `json:"about"`

	// [Required] JSON Schema specifies arguments, types, and additional context
	// to guide the LLM on command invokation.
	Schema json.RawMessage `json:"schema"`
}

// Foundational identity of LLMs
type LLM interface {
	// Model ID as defined by the vendor
	ModelID() string

	// Encode prompt to bytes:
	// - encoding prompt as prompt markup supported by LLM
	// - encoding prompt to envelop supported by LLM's hosting platform
	Encode([]Message, ...Opt) ([]byte, error)

	// Decode LLM's reply into pure text
	Decode([]byte) (Reply, error)
}
