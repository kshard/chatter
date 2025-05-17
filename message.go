//
// Copyright (C) 2025 Dmitry Kolesnikov
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

// Message is an abstract LLM I/O element
type Message interface {
	fmt.Stringer
	IsaMessage()
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

func (Reply) IsaMessage() {}

func (reply Reply) String() string {
	seq := make([]string, 0)
	for _, c := range reply.Content {
		switch v := (c).(type) {
		case Text:
			seq = append(seq, v.String())
		}
	}
	return strings.Join(seq, "")
}

// Invoke external tools
func (reply Reply) Invoke(f func(string, json.RawMessage) (json.RawMessage, error)) (Answer, error) {
	if reply.Stage != LLM_INVOKE {
		return Answer{}, nil
	}

	answer := Answer{Yield: make([]Json, 0)}
	for _, inv := range reply.Content {
		switch v := inv.(type) {
		case Invoke:
			val, err := f(v.Name, v.Args.Value)
			if err != nil {
				return answer, err
			}
			answer.Yield = append(answer.Yield, Json{ID: v.Args.ID, Source: v.Name, Value: val})
		}
	}

	return answer, nil
}

// Answer from external tools
type Answer struct {
	Yield []Json `json:"yield,omitempty"`
}

func (Answer) IsaMessage() {}

func (Answer) String() string { return "" }
