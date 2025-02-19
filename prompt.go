//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Prompt standardizes taxonomy of prompts for LLMs to solve complex tasks.
// See https://aclanthology.org/2023.findings-emnlp.946.pdf
//
// The container allows application to maintain semi-strucuted prompts while
// enabling efficient serialization into the textual prompt (aiming for quality).
// At the glance the prompt is structured:
//
//	 {task}. {guidelines}.
//		1. {requirements}
//		2. ...
//	 {feedback}
//	 {examples}
//	 {context}
//	 ...
//	 {input}
type Prompt struct {
	Task     string    `json:"task,omitempty"`
	Sections []Section `json:"sections,omitempty"`
}

// The task is a summary of what you want the prompt to do.
func (prompt *Prompt) WithTask(task string, args ...any) *Prompt {
	prompt.Task = fmt.Sprintf(Sentence(task), args...)
	return prompt
}

// Append section to prompt
func (prompt *Prompt) With(s Section) *Prompt {
	prompt.Sections = append(prompt.Sections, s)
	return prompt
}

// Helper function to make sequence of single prompt
func (prompt *Prompt) ToSeq() []fmt.Stringer { return []fmt.Stringer{prompt} }

// The prompt consisting of multiple sections, each is block of text.
type Section interface {
	fmt.Stringer
	Kind() Kind
}

// The type of the section.
type Kind int

const (
	TEXT Kind = iota
	RULES
	FEEDBACK
	EXAMPLE
	CONTEXT
	INPUT
	BLOB
)

// Snippet is the sequence to statements annotated with note for the model.
type Snippet struct {
	Type Kind     `json:"type,omitempty"`
	Note string   `json:"note,omitempty"`
	Text []string `json:"text,omitempty"`
}

func (r Snippet) Kind() Kind { return r.Type }

func (r Snippet) String() string {
	switch r.Type {
	case TEXT:
		return r.toText()
	case RULES:
		return r.toRules()
	case FEEDBACK:
		return r.toList()
	case CONTEXT:
		return r.toList()
	case INPUT:
		return r.toList()
	case BLOB:
		return r.toBlob()
	default:
		return r.toText()
	}
}

// convert snippet to block of text
func (r Snippet) toText() string {
	seq := make([]string, 0)
	if len(r.Note) > 0 {
		seq = append(seq, Sentence(r.Note))
	}
	for _, t := range r.Text {
		seq = append(seq, Sentence(t))
	}

	return strings.Join(seq, " ")
}

// convert shapless structure, dump as a raw file
func (r Snippet) toBlob() string {
	var sb strings.Builder
	if len(r.Note) > 0 {
		sb.WriteString(sentence(r.Note, ":"))
		sb.WriteString("\n")
	}
	for _, t := range r.Text {
		sb.WriteString(t)
		sb.WriteString("\n")
	}

	return sb.String()
}

// convert snippet to unordered list
func (r Snippet) toList() string {
	seq := make([]string, 0)
	if len(r.Note) > 0 {
		seq = append(seq, sentence(r.Note, ":"))
	}
	for _, t := range r.Text {
		seq = append(seq, "- "+Sentence(t))
	}

	return strings.Join(seq, "\n")
}

// convert snippet to ordered list
func (r Snippet) toRules() string {
	seq := make([]string, 0)
	if len(r.Note) > 0 {
		seq = append(seq, sentence(r.Note, ":"))
	}
	for i, t := range r.Text {
		seq = append(seq, strconv.Itoa(i+1)+". "+Sentence(t))
	}

	return strings.Join(seq, "\n")
}

// Examples how to complete the task, gives the input/output pair
type Example struct {
	Input string `json:"input,omitempty"`
	Reply string `json:"reply,omitempty"`
}

func (e Example) Kind() Kind { return EXAMPLE }

func (e Example) String() string {
	var sb strings.Builder
	sb.WriteString("Example Input: ")
	sb.WriteString(e.Input)
	sb.WriteString("\n")
	sb.WriteString("Expected Output: ")
	sb.WriteString(e.Reply)

	return sb.String()
}

// Guide LLM on how to complete the task.
//
//	prompt.With(
//		chatter.Guide(...)
//	)
func Guide(note string, text ...string) Snippet {
	s := Snippet{Type: TEXT, Note: sentence(note, "."), Text: make([]string, len(text))}
	for i, t := range text {
		s.Text[i] = Sentence(t)
	}
	return s
}

// Requirements is all about giving as much information as possible to ensure
// your response does not use any incorrect assumptions.
//
//	prompt.With(
//		chatter.Rules(...)
//	)
func Rules(note string, text ...string) Snippet {
	s := Snippet{Type: RULES, Note: sentence(note, ":"), Text: make([]string, len(text))}
	for i, t := range text {
		s.Text[i] = Sentence(t)
	}
	return s
}

// Give the feedback to LLM on previous completion of the task.
//
//	prompt.With(
//		chatter.Feedback(...)
//	)
func Feedback(note string, text ...string) Snippet {
	s := Snippet{Type: FEEDBACK, Note: sentence(note, ":"), Text: make([]string, len(text))}
	for i, t := range text {
		s.Text[i] = Sentence(t)
	}
	return s
}

// Additional information required to complete the task.
//
//	prompt.With(
//		chatter.Context(...)
//	)
func Context(note string, text ...string) Snippet {
	return Snippet{Type: CONTEXT, Note: sentence(note, ":"), Text: text}
}

// Input data required to complete the task.
//
//	prompt.With(
//		chatter.Input(...)
//	)
func Input(note string, text ...string) Snippet {
	return Snippet{Type: INPUT, Note: sentence(note, ":"), Text: text}
}

// Blob unformatted input data required to complete the task.
//
//	prompt.With(
//		chatter.Blob(...)
//	)
func Blob(note string, text ...string) Snippet {
	return Snippet{Type: BLOB, Note: sentence(note, ":"), Text: text}
}

// Ground level constrain of the model behavior.
// The latin meaning "something that has been laid down".
// Think about it as a cornerstone of the model behavior.
// "Act as <role>" ...
// Setting a specific role for a given prompt increases the likelihood of
// more accurate information, when done appropriately.
type Stratum string

func (s Stratum) String() string { return string(s) }

//------------------------------------------------------------------------------

// Helper function for making completition of sentences/pharses.
func Sentence(s string) string { return sentence(s, ".") }

var reSingleLine = regexp.MustCompile("[\r\n\t ]+")

const punctuations = ".:;,!?"

func sentence(s, dot string) string {
	s = reSingleLine.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return ""
	}

	c := s[len(s)-1]
	if !strings.ContainsRune(punctuations, rune(c)) {
		s = s + dot
	}

	return s
}

//------------------------------------------------------------------------------

// Converts prompt to string
func (prompt Prompt) String() string {
	var pb builder
	// {task}. {guide}.
	pb.text(&prompt)

	// 1. {rule}
	// 2. ...
	pb.unit(&prompt, RULES)

	// {feedback}
	pb.unit(&prompt, FEEDBACK)

	// Example Input: {input}
	// Expected Output: {output}
	pb.join(&prompt, EXAMPLE)

	// {about context}
	// - {context}
	// - {context}
	// - ...
	pb.join(&prompt, CONTEXT)

	//	{about input}:
	//	- {input}
	//	- {input}
	//	- ...
	//
	pb.unit(&prompt, INPUT)
	pb.join(&prompt, BLOB)

	return pb.String()
}

type builder struct{ seq []string }

func (b *builder) text(prompt *Prompt) {
	if b.seq == nil {
		b.seq = make([]string, 0)
	}

	var seq = make([]string, 0)
	if len(prompt.Task) != 0 {
		seq = append(seq, prompt.Task)
	}
	for _, t := range prompt.Sections {
		if t.Kind() == TEXT {
			seq = append(seq, t.String())
		}
	}

	txt := strings.Join(seq, " ")
	if len(txt) > 0 {
		b.seq = append(b.seq, txt)
	}
}

func (b *builder) unit(prompt *Prompt, kind Kind) {
	if b.seq == nil {
		b.seq = make([]string, 0)
	}

	var req string
	for _, t := range prompt.Sections {
		if t.Kind() == kind {
			if r, ok := t.(Snippet); ok {
				if len(req) != 0 {
					r.Note = ""
				} else {
					req = r.Note
				}
				b.seq = append(b.seq, r.String())
			} else {
				b.seq = append(b.seq, t.String())
			}
		}
	}
}

func (b *builder) join(prompt *Prompt, kind Kind) {
	if b.seq == nil {
		b.seq = make([]string, 0)
	}

	for _, t := range prompt.Sections {
		if t.Kind() == kind {
			b.seq = append(b.seq, t.String())
		}
	}
}

func (b *builder) String() string {
	return strings.Join(b.seq, "\n\n")
}
