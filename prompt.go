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

// Prompt data type consisting of context and bag of exchange messages.
type Prompt struct {
	// Ground level constrain of the model behavior.
	// The latin meaning "something that has been laid down".
	// Think about it as a cornerstone of the model behavior.
	// "Act as <role>" ...
	Role string `json:"stratum,omitempty"`

	// The task is a summary of what you want the prompt to do.
	Task string `json:"task,omitempty"`

	// Instructions informs model how to complete the task.
	// Examples of how it could go about tasks.
	Instructions *Remark `json:"instructions,omitempty"`

	// Requirements is all about giving as much information as possible to ensure
	// your response does not use any incorrect assumptions.
	Requirements *Remark `json:"requirements,omitempty"`

	// Input data required to complete the task.
	Input *Remark `json:"input,omitempty"`

	// Additional information required to complete the task.
	Context *Remark `json:"context,omitempty"`
}

// Remark is the sequence to statements annotated with note for the model.
type Remark struct {
	Note string   `json:"note,omitempty"`
	Text []string `json:"text,omitempty"`
}

// Setting a specific role for a given prompt increases the likelihood of
// more accurate information, when done appropriately.
func (prompt *Prompt) WithRole(role string) *Prompt {
	prompt.Role = strings.TrimSuffix(strings.TrimSpace(role), ".")
	return prompt
}

// The task is a summary of what you want the prompt to do.
func (prompt *Prompt) WithTask(task string, args ...any) *Prompt {
	prompt.Task = fmt.Sprintf(
		strings.TrimSuffix(strings.TrimSpace(task), "."),
		args...,
	)
	return prompt
}

// Instructions informs model how to complete the task.
// Examples of how it could go about tasks.
func (prompt *Prompt) WithInstruction(ins string, args ...any) *Prompt {
	if prompt.Instructions == nil {
		prompt.Instructions = &Remark{Note: "", Text: []string{}}
	}

	prompt.Instructions.Text = append(
		prompt.Instructions.Text,
		fmt.Sprintf(strings.TrimSpace(ins), args...),
	)
	return prompt
}

// Requirements is all about giving as much information as possible to ensure
// your response does not use any incorrect assumptions.
func (prompt *Prompt) WithRequirements(note string) *Prompt {
	if prompt.Requirements == nil {
		prompt.Requirements = &Remark{Note: "", Text: []string{}}
	}

	prompt.Requirements.Note = strings.TrimSpace(note)
	return prompt
}

// Requirements is all about giving as much information as possible to ensure
// your response does not use any incorrect assumptions.
func (prompt *Prompt) WithRequirement(req string, args ...any) *Prompt {
	if prompt.Requirements == nil {
		prompt.Requirements = &Remark{Note: "", Text: []string{}}
	}

	prompt.Requirements.Text = append(prompt.Requirements.Text,
		fmt.Sprintf(strings.TrimSpace(req), args...),
	)
	return prompt
}

// Input data required to complete the task.
func (prompt *Prompt) WithInput(about string, input []string) *Prompt {
	prompt.Input = &Remark{
		Note: strings.TrimSuffix(strings.TrimSpace(about), ":"),
		Text: input,
	}

	return prompt
}

// Additional information required to complete the task.
func (prompt *Prompt) WithContext(about string, context []string) *Prompt {
	prompt.Context = &Remark{
		Note: strings.TrimSuffix(strings.TrimSpace(about), ":"),
		Text: context,
	}

	return prompt
}

// Prompt to string formatter
type Formatter interface {
	ToString(*strings.Builder, *Prompt)
}

// Generic prompt formatter. Build prompt following the best approach
//
//	{role}. {task}. {instructions}.
//	1. {requirements}
//	2. {requirements}
//	3. ...
//
//	{about input}:
//	- {input}
//	- {input}
//	- ...
//
//	{about context}
//	- {context}
//	- {context}
//	- ...
func NewFormatter(role string) Formatter {
	return defaultFormatter{role: strings.TrimSpace(role)}
}

type defaultFormatter struct {
	// Ground level constrain of the model behavior.
	// The latin meaning "something that has been laid down".
	// Think about it as a cornerstone of the model behavior.
	// "Act as <stratum>" ...
	role string
}

func (p defaultFormatter) ToString(sb *strings.Builder, prompt *Prompt) {
	if len(prompt.Role) == 0 {
		prompt.WithRole(p.role)
	}

	if len(prompt.Role) > 0 {
		sb.WriteString(prompt.Role)
		sb.WriteString(". ")
	}

	if len(prompt.Task) > 0 {
		sb.WriteString(asSingleLine(prompt.Task))
		sb.WriteString(". ")
	}

	p.pp(sb, prompt.Instructions)
	p.ol(sb, prompt.Requirements)
	p.ul(sb, prompt.Input)
	p.ul(sb, prompt.Context)
}

// write remark as sequence of sentences
func (p defaultFormatter) pp(sb *strings.Builder, remark *Remark) {
	if remark == nil || len(remark.Text) == 0 {
		return
	}

	if len(remark.Text) == 1 {
		sb.WriteString(strings.TrimSuffix(asSingleLine(remark.Text[0]), "."))
		sb.WriteString(". ")
		return
	}

	for _, t := range remark.Text {
		sb.WriteString(strings.TrimSuffix(asSingleLine(t), "."))
		sb.WriteString(". ")
	}
}

// write remark as unordered list
func (p defaultFormatter) ul(sb *strings.Builder, remark *Remark) {
	if remark == nil || len(remark.Text) == 0 {
		return
	}

	if len(remark.Text) == 1 {
		sb.WriteString(asSingleLine(remark.Text[0]))
		sb.WriteString("\n")
		return
	}

	sb.WriteString("\n")
	sb.WriteString(asSingleLine(remark.Note))
	sb.WriteString("\n")

	for _, t := range remark.Text {
		sb.WriteString("* ")
		sb.WriteString(asSingleLine(t))
		sb.WriteString("\n")
	}
}

// write remark as ordered list
func (p defaultFormatter) ol(sb *strings.Builder, remark *Remark) {
	if remark == nil || len(remark.Text) == 0 {
		return
	}

	if len(remark.Text) == 1 {
		sb.WriteString(asSingleLine(remark.Text[0]))
		sb.WriteString("\n")
		return
	}

	sb.WriteString("\n")
	sb.WriteString(asSingleLine(remark.Note))
	sb.WriteString("\n")

	for i, t := range remark.Text {
		sb.WriteString(strconv.Itoa(i + 1))
		sb.WriteString(". ")
		sb.WriteString(asSingleLine(t))
		sb.WriteString("\n")
	}
}

var reSingleLine = regexp.MustCompile("[\r\n\t ]+")

func asSingleLine(s string) string {
	return reSingleLine.ReplaceAllString(s, " ")
}
