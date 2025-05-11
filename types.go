//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

// Stage of the interaction with LLM
type Stage string

const (
	// LLM has a result to return
	LLM_RETURN = Stage("return")

	// LLM has a result to return but it was truncated (e.g. max tokens, stop sequence)
	LLM_INCOMPLETE = Stage("incomplete")

	// LLM requires to invoke external command/tools
	LLM_INVOKE = Stage("invoke")

	// LLM has aborted execution due to error
	LLM_ERROR = Stage("error")
)

// The reply from LLMs
type Reply struct {
	Stage   Stage     `json:"stage"`
	Usage   Usage     `json:"usage"`
	Content []Content `json:"content"`
}

func (reply Reply) String() string {
	var sb strings.Builder
	for _, block := range reply.Content {
		switch v := block.(type) {
		case ContentText:
			sb.WriteString(v.Text)
		}
	}
	return sb.String()
}

// Invoke external tools
func (reply Reply) Invoke(f func(string, json.RawMessage) (json.RawMessage, error)) (Answer, error) {
	if reply.Stage != LLM_INVOKE {
		return Answer{}, nil
	}

	answer := Answer{Yield: make([]ContentJson, 0)}
	for _, inv := range reply.Content {
		switch v := inv.(type) {
		case Invoke:
			val, err := f(v.Name, v.Args.Value)
			if err != nil {
				return answer, err
			}
			answer.Yield = append(answer.Yield, ContentJson{Source: v.Args.Source, Value: val})
		}
	}

	return answer, nil
}

// Return of external tool / command
type Answer struct {
	Yield []ContentJson
}

func (r Answer) String() string {
	seq := make([]string, len(r.Yield))
	for i, yield := range r.Yield {
		seq[i] = yield.Source
	}
	return strings.Join(seq, ", ")
}

// LLM Usage stats
type Usage struct {
	InputTokens int `json:"inputTokens"`
	ReplyTokens int `json:"replyTokens"`
}

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

//
// Content blocks
//

// Block of Content in LLMs Messages
type Content interface{ HKT1(Content) }

type ContentText struct {
	Text string `json:"text"`
}

func (ContentText) HKT1(Content) {}

type ContentJson struct {
	Source string          `json:"source"`
	Value  json.RawMessage `json:"value"`
}

func (ContentJson) HKT1(Content) {}

// LLM invokes external tool / command
type Invoke struct {
	Name    string      `json:"name"`
	Args    ContentJson `json:"args"`
	Message any         `json:"message"`
}

func (Invoke) HKT1(Content) {}

func (inv Invoke) RawMessage() any { return inv.Message }
