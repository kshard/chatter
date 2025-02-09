//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

import (
	"fmt"
	"testing"

	"github.com/fogfish/it/v2"
)

func TestPromptWithRole(t *testing.T) {
	v := "assistant"
	p := &Prompt{}
	p.WithRole(v)
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Equal(p.Role, v),
		it.Equal(string(txt), v+"."),
	)
}

func TestPromptWithTask(t *testing.T) {
	v := "Translate the following text"
	p := &Prompt{}
	p.WithTask(v)
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Equal(p.Task, v),
		it.Equal(string(txt), v+"."),
	)
}

func TestPromptWithInstruction(t *testing.T) {
	v := "Use formal language"
	p := &Prompt{}
	p.WithInstruction(v)
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Seq(p.Instructions.Text).Equal(v),
		it.Equal(string(txt), v+"."),
	)
}

func TestPromptWithRequirements(t *testing.T) {
	v := "Ensure accuracy"
	p := &Prompt{}
	p.WithRequirements(v)
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Equal(p.Requirements.Note, v),
		it.Equal(string(txt), ""),
	)
}

func TestPrompt_WithRequirement(t *testing.T) {
	v := "Ensure accuracy"
	y := "Translate all text"
	p := &Prompt{}
	p.WithRequirements(v)
	p.WithRequirement(y)
	p.WithRequirement(y)
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Equal(p.Requirements.Note, v),
		it.Seq(p.Requirements.Text).Equal(y, y),
		it.Equal(string(txt), fmt.Sprintf("%s\n1. %s\n2. %s", v, y, y)),
	)
}

func TestPromptWithExample(t *testing.T) {
	v := "Hello"
	y := "Hola"
	p := &Prompt{}
	p.WithExample(v, y)
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Seq(p.Examples).Equal(Example{Input: v, Output: y}),
		it.Equal(string(txt), fmt.Sprintf("Example\nInput: %s\nOutput: %s", v, y)),
	)
}

func TestPromptWithInput(t *testing.T) {
	a := "Translate the following"
	h := "Hello"
	w := "World"
	p := &Prompt{}
	p.WithInput(a, []string{h, w})
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Seq(p.Input.Text).Equal(h, w),
		it.Equal(string(txt), fmt.Sprintf("%s\n* %s\n* %s", a, h, w)),
	)
}

func TestPromptWithContext(t *testing.T) {
	a := "Context information"
	h := "Hello"
	w := "World"
	p := &Prompt{}
	p.WithContext(a, []string{h, w})
	txt, _ := p.MarshalText()

	it.Then(t).Should(
		it.Seq(p.Context.Text).Equal(h, w),
		it.Equal(string(txt), fmt.Sprintf("%s\n* %s\n* %s", a, h, w)),
	)
}
