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
		it.Equal(string(p.Task), e),
		it.Equal(p.String(), e),
	)
}

func TestPromptWithGuide(t *testing.T) {
	v := "Use formal language"
	e := v + "."

	p := &Prompt{}
	p.WithGuide("", v)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Guide{Note: "", Text: []string{e}}),
		it.Equal(p.String(), e),
	)
}

func TestPromptWithRules1(t *testing.T) {
	v := "Ensure accuracy"
	e := "1. " + v + "."

	p := &Prompt{}
	p.WithRules("", v)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Rules{Note: "", Text: []string{v + "."}}),
		it.Equal(p.String(), e),
	)
}

func TestPromptWithRules2(t *testing.T) {
	v := "Ensure accuracy"
	y := "Translate all text"

	p := &Prompt{}
	p.WithRules(v, y, y)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Rules{Note: v + ":", Text: []string{y + ".", y + "."}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n1. %s.\n2. %s.", v, y, y)),
	)
}

func TestPromptWithExample(t *testing.T) {
	v := "Hello"
	y := "Hola"

	p := &Prompt{}
	p.WithExample(v, y)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Example{Input: v, Reply: y}),
		it.Equal(p.String(), fmt.Sprintf("Example Input:\n%s\nExpected Output:\n%s\n\n", v, y)),
	)
}

func TestPromptWithInput(t *testing.T) {
	a := "Translate the following"
	h := "Hello"
	w := "World"

	p := &Prompt{}
	p.WithInput(a, h, w)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Input{Note: a + ":", Text: []string{h, w}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n- %s.\n- %s.", a, h, w)),
	)
}

func TestPromptWithAttach(t *testing.T) {
	a := "Translate the following"
	h := "Hello World"

	p := &Prompt{}
	p.WithBlob(a, h)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Blob{Note: a + ":", Text: h}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n%s\n", a, h)),
	)
}

func TestPromptWithContext(t *testing.T) {
	a := "Context information"
	h := "Hello"
	w := "World"

	p := &Prompt{}
	p.WithContext(a, h, w)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Context{Note: a + ":", Text: []string{h, w}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n- %s.\n- %s.", a, h, w)),
	)
}

func TestPromptWithFeedback(t *testing.T) {
	a := "Context information"
	h := "Hello"
	w := "World"

	p := &Prompt{}
	p.WithFeedback(a, h, w)

	it.Then(t).Should(
		it.Seq(p.Content).Equal(Feedback{Note: a + ":", Text: []string{h + ".", w + "."}}),
		it.Equal(p.String(), fmt.Sprintf("%s:\n- %s.\n- %s.", a, h, w)),
	)
}
