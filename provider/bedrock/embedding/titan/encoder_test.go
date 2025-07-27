//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package titan

import (
	"encoding/json"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func TestEncoderTextInput(t *testing.T) {
	f, err := factory(EMBEDDING_256)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Hello world"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "Hello world",
		"dimensions": 256
	}`))
}

func TestEncoderPromptInput(t *testing.T) {
	f, err := factory(EMBEDDING_256)()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Summarize the following text")
	prompt.WithInput("Text to summarize:", "The quick brown fox jumps over the lazy dog")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "regex:Summarize the following text.\nText to summarize:\n- The quick brown fox jumps over the lazy dog.",
		"dimensions": 256
	}`))
}

func TestEncoderMultipleTextInputs(t *testing.T) {
	f, err := factory(EMBEDDING_256)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("First part "))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("second part "))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("third part"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "First part second part third part",
		"dimensions": 256
	}`))
}

func TestEncoderNoOpMethods(t *testing.T) {
	f, err := factory(EMBEDDING_256)()
	it.Then(t).Must(it.Nil(err))

	// Test that no-op methods don't affect the output
	f.WithInferrer(provider.Inferrer{
		Temperature: 0.7,
		TopP:        0.9,
		MaxTokens:   1000,
	})

	f.WithCommand(chatter.Cmd{
		Cmd:    "test",
		About:  "Test command",
		Schema: json.RawMessage(`{"type": "object"}`),
	})

	err = f.AsStratum(chatter.Stratum("System message"))
	it.Then(t).Must(it.Nil(err))

	err = f.AsAnswer(&chatter.Answer{
		Yield: []chatter.Json{
			{
				ID:     "tool-1",
				Source: "calculator",
				Value:  json.RawMessage(`{"result": 42}`),
			},
		},
	})
	it.Then(t).Must(it.Nil(err))

	err = f.AsReply(&chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("Assistant response"),
		},
	})
	it.Then(t).Must(it.Nil(err))

	// Only add actual text that should appear
	err = f.AsText(chatter.Text("Actual text content"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "Actual text content"
	}`))
}

func TestEncoderEmptyInput(t *testing.T) {
	f, err := factory(EMBEDDING_256)()
	it.Then(t).Must(it.Nil(err))

	// Don't add any content, just build
	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "",
		"dimensions": 256
	}`))
}

func TestEncoderSpecialCharacters(t *testing.T) {
	f, err := factory(EMBEDDING_256)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Text with \"quotes\" and \n newlines \t tabs"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "regex:Text with \\\"quotes\\\" and \\s+ newlines \\s+ tabs"
	}`))
}
