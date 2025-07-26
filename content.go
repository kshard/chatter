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

// Content is the core building block for I/O with LLMs.
// It defines either input prompt or result of the LLM execution.
// For example,
//   - Prompt is a content either simple plain [Text] or semistructured [Prompt].
//   - LLM replies with generated [Text], [Vector] or [Invoke] instructions.
//   - Invocation of external tools is orchestrated using [Json] content.
//   - etc.
//
// The content itself is encapsulated in sequence of [Message] forming a conversation.
type Content interface{ fmt.Stringer }

// Text is a plain text either part of prompt or LLM's reply.
// For simplicity of library's api, text is also representing
// a [Message] (HKT1(Message)), allowing it to be used directly as input (prompt) to LLM
type Text string

func (t Text) HKT1(Message)   {}
func (t Text) String() string { return string(t) }

func (t Text) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Text string `json:"text,omitempty"`
	}{Text: string(t)})
}

// Json is a structured object (JSON object) that can be used as input to LLMs
// or as a reply from LLMs.
//
// Json is a key abstraction for LLMs integration with external tools.
// It is used to pass structured data from LLM to the tool and vice versa,
// supporting invocation and answering the resuls.
type Json struct {
	// Unique identifier of Json objects, used for tracking in the conversation
	// and correlating the input with output (invocations with answers).
	ID string `json:"id,omitempty"`

	// Unique identifier of the source of the Json object.
	// For example, it can be a name of the tool that produced the output.
	Source string `json:"source,omitempty"`

	// Value of JSON Object
	Value json.RawMessage `json:"bag,omitempty"`
}

func (j Json) String() string {
	return string(j.Value)
}

//------------------------------------------------------------------------------

// Task is part of the [Prompt] that defines the task to be solved by LLM.
type Task string

func (t Task) HKT1(Message)   {}
func (t Task) String() string { return string(t) }

// Guide is part of the [Prompt] that guides LLM on how to complete the task.
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

// Rules is part of the [Prompt] that defines the rules and requirements to be
// followed by LLM. Use it to give as much information as possible to ensure
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

// Feedback is part of the [Prompt] that gives feedback to LLM on previous
// completion of the task (e.g. errors).
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

// Example is part of the [Prompt] that gives examples how to complete the task.
type Example struct {
	Input string `json:"input,omitempty"`
	Reply string `json:"reply,omitempty"`
}

func (e Example) String() string {
	return fmt.Sprintf("Example Input:\n%s\nExpected Output:\n%s\n\n", e.Input, e.Reply)
}

// Context is part of the [Prompt] that provides additional information
// required to complete the task.
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

// Input is part of the [Prompt] that provides input data required to
// complete the task.
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

// Blob is part of the [Prompt] that provides unformatted input data required to
// complete the task.
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

//------------------------------------------------------------------------------

// Invoke is a special content bloc defining interaction with external functions.
// Invoke is generated by LLMs when execution of external tools is required.
//
// It is expected that client code will use [Reply.Invoke] to process
// the invocation and call the function with the name and arguments.
//
// [Answer] is returned with the results of the function call.
type Invoke struct {
	// Unique identifier of the tool model wants to use.
	// The name is used to lookup the tool in the registry.
	Cmd string `json:"name"`

	// Arguments to the tool, which are passed as a JSON object.
	Args Json `json:"args"`

	// Original LLM message that triggered the invocation, as defined by the providers API.
	// The message is used to maintain the converstation history and context.
	Message any `json:"-"`
}

func (inv Invoke) String() string  { return fmt.Sprintf("invoke @%s", inv.Cmd) }
func (inv Invoke) RawMessage() any { return inv.Message }

//------------------------------------------------------------------------------

// Vector is a sequence of float32 numbers representing the embedding vector.
type Vector []float32

func (v Vector) String() string { return fmt.Sprintf("%v", []float32(v)) }

func (v Vector) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Vector []float32 `json:"vector,omitempty"`
	}{Vector: []float32(v)})
}
