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

// func TestPromptWithRole(t *testing.T) {
// 	v := "assistant"
// 	p := &Prompt{}
// 	p.WithRole(v)
// 	txt, _ := p.MarshalText()

// 	it.Then(t).Should(
// 		it.Equal(p.Role, v),
// 		it.Equal(string(txt), v+"."),
// 	)
// }

func TestPromptWithTask(t *testing.T) {
	v := "Translate the following text"
	e := v + "."

	p := &Prompt{}
	p.WithTask(v)

	it.Then(t).Should(
		it.Equal(p.Task, e),
		it.Equal(p.String(), e),
	)
}

func TestPromptWithGuide(t *testing.T) {
	v := "Use formal language"
	e := v + "."

	p := &Prompt{}
	p.With(Guide("", v))

	it.Then(t).Should(
		it.Seq(p.Sections).Equal(Snippet{Type: TEXT, Note: "", Text: []string{e}}),
		it.Equal(p.String(), e),
	)
}

func TestPromptWithRules1(t *testing.T) {
	v := "Ensure accuracy"
	e := "1. " + v + "."

	p := &Prompt{}
	p.With(Rules("", v))

	it.Then(t).Should(
		it.Seq(p.Sections).Equal(Snippet{Type: RULES, Note: "", Text: []string{v + "."}}),
		it.Equal(p.String(), e),
	)
}

func TestPromptWithRules2(t *testing.T) {
	v := "Ensure accuracy"
	y := "Translate all text"

	p := &Prompt{}
	p.With(Rules(v, y, y))

	it.Then(t).Should(
		it.Seq(p.Sections).Equal(Snippet{Type: RULES, Note: v + ":", Text: []string{y + ".", y + "."}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n1. %s.\n2. %s.", v, y, y)),
	)
}

func TestPromptWithExample(t *testing.T) {
	v := "Hello"
	y := "Hola"

	p := &Prompt{}
	p.With(Example{Input: v, Reply: y})

	it.Then(t).Should(
		it.Seq(p.Sections).Equal(Example{Input: v, Reply: y}),
		it.Equal(p.String(), fmt.Sprintf("Example Input: %s\nExpected Output: %s", v, y)),
	)
}

func TestPromptWithInput(t *testing.T) {
	a := "Translate the following"
	h := "Hello"
	w := "World"

	p := &Prompt{}
	p.With(Input(a, h, w))

	it.Then(t).Should(
		it.Seq(p.Sections).Equal(Snippet{Type: INPUT, Note: a + ":", Text: []string{h, w}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n- %s.\n- %s.", a, h, w)),
	)
}

func TestPromptWithContext(t *testing.T) {
	a := "Context information"
	h := "Hello"
	w := "World"

	p := &Prompt{}
	p.With(Context(a, h, w))

	it.Then(t).Should(
		it.Seq(p.Sections).Equal(Snippet{Type: CONTEXT, Note: a + ":", Text: []string{h, w}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n- %s.\n- %s.", a, h, w)),
	)
}

func TestPromptWithFeedback(t *testing.T) {
	a := "Context information"
	h := "Hello"
	w := "World"

	p := &Prompt{}
	p.With(Feedback(a, h, w))

	it.Then(t).Should(
		it.Seq(p.Sections).Equal(Snippet{Type: FEEDBACK, Note: a + ":", Text: []string{h + ".", w + "."}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n- %s.\n- %s.", a, h, w)),
	)
}
