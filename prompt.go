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
	"strings"
)

// Ground level constrain of the model behavior.
// The latin meaning "something that has been laid down".
// Think about it as a cornerstone of the model behavior.
// "Act as <role>" ...
// Setting a specific role for a given prompt increases the likelihood of
// more accurate information, when done appropriately.
type Stratum string

// Stratum is LLM Message
func (Stratum) IsaMessage() {}

func (s Stratum) String() string { return string(s) }

//------------------------------------------------------------------------------

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
	Task    Task      `json:"task,omitempty"`
	Content []Content `json:"content,omitempty"`
}

var (
	_ Message = (*Prompt)(nil)
)

// Prompt is LLM Message
func (*Prompt) IsaMessage() {}

// Add Content block into LLM's prompt
func (prompt *Prompt) With(block Content) *Prompt {
	prompt.Content = append(prompt.Content, block)
	return prompt
}

// The task is a summary of what you want the prompt to do.
//
//	prompt.WithTask(...)
func (prompt *Prompt) WithTask(task string, args ...any) *Prompt {
	prompt.Task = Task(fmt.Sprintf(Sentence(task), args...))
	return prompt
}

// Guide LLM on how to complete the task.
//
//	prompt.WithGuid(...)
func (prompt *Prompt) WithGuide(note string, text ...string) *Prompt {
	guide := Guide{
		Note: sentence(note, "."),
		Text: make([]string, len(text)),
	}
	for i, t := range text {
		guide.Text[i] = Sentence(t)
	}

	prompt.Content = append(prompt.Content, guide)
	return prompt
}

// Requirements is all about giving as much information as possible to ensure
// your response does not use any incorrect assumptions.
//
//	prompt.WithRules(...)
func (prompt *Prompt) WithRules(note string, text ...string) *Prompt {
	rules := Rules{
		Note: sentence(note, ":"),
		Text: make([]string, len(text)),
	}
	for i, t := range text {
		rules.Text[i] = Sentence(t)
	}

	prompt.Content = append(prompt.Content, rules)
	return prompt
}

// Give the feedback to LLM on previous completion of the task.
//
//	prompt.WithFeedback(...)
func (prompt *Prompt) WithFeedback(note string, text ...string) *Prompt {
	feedback := Feedback{
		Note: sentence(note, ":"),
		Text: make([]string, len(text)),
	}
	for i, t := range text {
		feedback.Text[i] = Sentence(t)
	}

	prompt.Content = append(prompt.Content, feedback)
	return prompt
}

// Give examples to LLM about input data and expected outcomes.
//
//	prompt.WithExample(...)
func (prompt *Prompt) WithExample(input, reply string) *Prompt {
	example := Example{Input: input, Reply: reply}

	prompt.Content = append(prompt.Content, example)
	return prompt
}

// Additional information required to complete the task.
//
//	prompt.WithContext(...)
func (prompt *Prompt) WithContext(note string, text ...string) *Prompt {
	context := Context{
		Note: sentence(note, ":"),
		Text: text,
	}

	prompt.Content = append(prompt.Content, context)
	return prompt
}

// Input data required to complete the task.
//
//	prompt.WithInput(...)
func (prompt *Prompt) WithInput(note string, text ...string) *Prompt {
	input := Input{
		Note: sentence(note, ":"),
		Text: text,
	}

	prompt.Content = append(prompt.Content, input)
	return prompt
}

// Blob unformatted input data required to complete the task.
//
//	prompt.WithBlob(...)
func (prompt *Prompt) WithBlob(note string, text string) *Prompt {
	blob := Blob{
		Note: sentence(note, ":"),
		Text: text,
	}

	prompt.Content = append(prompt.Content, blob)
	return prompt
}

// Helper function to make sequence of single prompt
func (prompt *Prompt) ToSeq() []Message { return []Message{prompt} }

// Converts prompt to structured string
func (prompt *Prompt) String() string {
	seq := make([]string, 0)

	//	 {task}. {guidelines}.
	if len(prompt.Task) > 0 {
		seq = append(seq, prompt.Task.String())
	}
	for _, x := range filter[Guide](prompt) {
		seq = append(seq, x.String())
	}

	//		1. {requirements}
	//		2. ...
	for _, x := range filter[Rules](prompt) {
		seq = append(seq, x.String())
	}

	//	 {feedback}
	for _, x := range filter[Feedback](prompt) {
		seq = append(seq, x.String())
	}

	//	 {examples}
	for _, x := range filter[Example](prompt) {
		seq = append(seq, x.String())
	}

	//	 {context}
	for _, x := range filter[Context](prompt) {
		seq = append(seq, x.String())
	}

	//	 ...
	//	 {input}
	for _, x := range filter[Input](prompt) {
		seq = append(seq, x.String())
	}
	for _, x := range filter[Blob](prompt) {
		seq = append(seq, x.String())
	}

	return strings.Join(seq, "\n")
}

// helper function to filter
func filter[T Content](prompt *Prompt) []T {
	seq := make([]T, 0)
	for _, x := range prompt.Content {
		switch v := x.(type) {
		case T:
			seq = append(seq, v)
		}
	}

	return seq
}

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
