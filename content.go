//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//
// Content blocks
//

// Content is building block for the message
type Content interface{ fmt.Stringer }

// Plain text
type Text string

func (t Text) HKT1(Message)   {}
func (t Text) String() string { return string(t) }

// Json object
type Json struct {
	ID     string          `json:"id,omitempty"`
	Source string          `json:"source,omitempty"`
	Value  json.RawMessage `json:"bag,omitempty"`
}

func (j Json) String() string {
	return string(j.Value)
}

//------------------------------------------------------------------------------

// Invoke is a special content bloc defining interaction with external functions
type Invoke struct {
	Name    string `json:"name"`
	Args    Json   `json:"args"`
	Message any    `json:"-"`
}

func (inv Invoke) String() string  { return fmt.Sprintf("invoke @%s", inv.Name) }
func (inv Invoke) RawMessage() any { return inv.Message }

//------------------------------------------------------------------------------

// Text content
type Task string

func (t Task) HKT1(Message)   {}
func (t Task) String() string { return string(t) }

// Guide LLM on how to complete the task.
type Guide struct {
	Note string   `json:"note,omitempty"`
	Text []string `json:"guide,omitempty"`
}

func (g Guide) String() string {
	seq := make([]string, 0)
	if len(g.Note) > 0 {
		seq = append(seq, Sentence(g.Note))
	}
	for _, t := range g.Text {
		seq = append(seq, Sentence(t))
	}

	return strings.Join(seq, "\n")
}

// Requirements is all about giving as much information as possible to ensure
// your response does not use any incorrect assumptions.
type Rules struct {
	Note string   `json:"note,omitempty"`
	Text []string `json:"rules,omitempty"`
}

func (r Rules) String() string {
	seq := make([]string, 0)
	if len(r.Note) > 0 {
		seq = append(seq, sentence(r.Note, ":"))
	}
	for i, t := range r.Text {
		seq = append(seq, strconv.Itoa(i+1)+". "+Sentence(t))
	}

	return strings.Join(seq, "\n")
}

// Give the feedback to LLM on previous completion of the task.
type Feedback struct {
	Note string   `json:"note,omitempty"`
	Text []string `json:"feedback,omitempty"`
}

func (f Feedback) String() string {
	seq := make([]string, 0)
	if len(f.Note) > 0 {
		seq = append(seq, sentence(f.Note, ":"))
	}
	for _, t := range f.Text {
		seq = append(seq, "- "+Sentence(t))
	}

	return strings.Join(seq, "\n")
}

func (f Feedback) Error() string { return f.String() }

// Examples how to complete the task, gives the input/output pair
type Example struct {
	Input string `json:"input,omitempty"`
	Reply string `json:"reply,omitempty"`
}

func (e Example) String() string {
	return fmt.Sprintf("Example Input:\n%s\nExpected Output:\n%s\n\n", e.Input, e.Reply)
}

// Additional information required to complete the task.
type Context struct {
	Note string   `json:"note,omitempty"`
	Text []string `json:"context,omitempty"`
}

func (c Context) String() string {
	seq := make([]string, 0)
	if len(c.Note) > 0 {
		seq = append(seq, sentence(c.Note, ":"))
	}
	for _, t := range c.Text {
		seq = append(seq, "- "+Sentence(t))
	}

	return strings.Join(seq, "\n")
}

// Input data required to complete the task.
type Input struct {
	Note string   `json:"note,omitempty"`
	Text []string `json:"input,omitempty"`
}

func (i Input) String() string {
	seq := make([]string, 0)
	if len(i.Note) > 0 {
		seq = append(seq, sentence(i.Note, ":"))
	}
	for _, t := range i.Text {
		seq = append(seq, "- "+Sentence(t))
	}

	return strings.Join(seq, "\n")
}

// Blob unformatted input data required to complete the task.
type Blob struct {
	Note string `json:"text,omitempty"`
	Text string `json:"blob,omitempty"`
}

func (b Blob) String() string {
	var sb strings.Builder
	if len(b.Note) > 0 {
		sb.WriteString(sentence(b.Note, ":"))
		sb.WriteString("\n")
	}
	sb.WriteString(b.Text)
	sb.WriteString("\n")

	return sb.String()
}
