//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llama

import (
	"encoding/json"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func TestEncoderInferenceConfiguration(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature: 0.7,
		TopP:        0.9,
		MaxTokens:   512,
	})

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"prompt": "<|begin_of_text|>",
		"temperature": 0.7,
		"top_p": 0.9,
		"max_gen_len": 512
	}`))
}

func TestEncoderSystemMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsStratum(chatter.Stratum("You are a helpful assistant"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"prompt": "regex:<\\|begin_of_text\\|>\\n<\\|start_header_id\\|>system<\\|end_header_id\\|>\\nYou are a helpful assistant\\n<\\|eot_id\\|>\\n"
	}`))
}

func TestEncoderUserMessages(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Hello, how are you?"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"prompt": "regex:<\\|begin_of_text\\|>\\n<\\|start_header_id\\|>user<\\|end_header_id\\|>\\nHello, how are you\\?\\n<\\|eot_id\\|>\\n\\n<\\|start_header_id\\|>assistant<\\|end_header_id\\|>\\n"
	}`))
}

func TestEncoderPromptMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Translate the following text to Spanish")
	prompt.WithInput("Text to translate:", "Hello world")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"prompt": "regex:<\\|begin_of_text\\|>\\n<\\|start_header_id\\|>user<\\|end_header_id\\|>\\nTranslate the following text to Spanish\\.\\nText to translate:\\n- Hello world\\.\\n<\\|eot_id\\|>\\n\\n<\\|start_header_id\\|>assistant<\\|end_header_id\\|>\\n"
	}`))
}

func TestEncoderAnswerMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	// Test with empty yield (should be no-op)
	err = f.AsAnswer(&chatter.Answer{})
	it.Then(t).Must(it.Nil(err))

	// Test with actual yield
	jsonData := `{"result": "success"}`
	answer := &chatter.Answer{
		Yield: []chatter.Json{
			{
				ID:     "tool-1",
				Source: "calculator",
				Value:  json.RawMessage(jsonData),
			},
		},
	}

	err = f.AsAnswer(answer)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"prompt": "regex:<\\|begin_of_text\\|>\\n<\\|start_header_id\\|>user<\\|end_header_id\\|>\\n.*\"result\":\\s*\"success\".*\\n<\\|eot_id\\|>\\n\\n<\\|start_header_id\\|>assistant<\\|end_header_id\\|>\\n"
	}`))
}

func TestEncoderReplyMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	reply := &chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("This is a response from the assistant"),
		},
	}

	err = f.AsReply(reply)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"prompt": "regex:<\\|begin_of_text\\|>This is a response from the assistant\\n<\\|eot_id\\|>\\n"
	}`))
}

func TestEncoderConversationFlow(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature: 0.5,
		MaxTokens:   256,
	})

	// System message
	err = f.AsStratum(chatter.Stratum("You are a helpful assistant"))
	it.Then(t).Must(it.Nil(err))

	// User message
	err = f.AsText(chatter.Text("What is 2+2?"))
	it.Then(t).Must(it.Nil(err))

	// Assistant reply
	reply := &chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("2+2 equals 4"),
		},
	}
	err = f.AsReply(reply)
	it.Then(t).Must(it.Nil(err))

	// Another user message
	err = f.AsText(chatter.Text("Thanks!"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"prompt": "regex:<\\|begin_of_text\\|>\\n<\\|start_header_id\\|>system<\\|end_header_id\\|>\\nYou are a helpful assistant\\n<\\|eot_id\\|>\\n\\n<\\|start_header_id\\|>user<\\|end_header_id\\|>\\nWhat is 2\\+2\\?\\n<\\|eot_id\\|>\\n\\n<\\|start_header_id\\|>assistant<\\|end_header_id\\|>\\n2\\+2 equals 4\\n<\\|eot_id\\|>\\n\\n<\\|start_header_id\\|>user<\\|end_header_id\\|>\\nThanks!\\n<\\|eot_id\\|>\\n\\n<\\|start_header_id\\|>assistant<\\|end_header_id\\|>\\n",
		"temperature": 0.5,
		"max_gen_len": 256
	}`))
}
