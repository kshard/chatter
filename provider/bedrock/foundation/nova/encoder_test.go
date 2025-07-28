//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package nova

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
		Temperature:   0.8,
		TopP:          0.95,
		MaxTokens:     1024,
		StopSequences: []string{"STOP", "END"},
	})

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"messages": [],
		"inferenceConfig": {
			"maxTokens": 1024,
			"temperature": 0.8,
			"topP": 0.95,
			"stopSequences": ["STOP", "END"]
		}
	}`))
}

func TestEncoderSystemMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsStratum(chatter.Stratum("You are a helpful AI assistant specialized in code analysis"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"system": [
			{
				"text": "You are a helpful AI assistant specialized in code analysis"
			}
		],
		"messages": []
	}`))
}

func TestEncoderUserMessages(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Explain quantum computing in simple terms"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"text": "Explain quantum computing in simple terms"
					}
				]
			}
		]
	}`))
}

func TestEncoderPromptMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Review the following code for security vulnerabilities")
	prompt.WithInput("Code snippet:", "SELECT")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"text": "regex:Review the following code for security vulnerabilities.\\nCode snippet:\\n- SELECT."
					}
				]
			}
		]
	}`))
}

func TestEncoderAnswerMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	// Test with empty yield (should be no-op)
	err = f.AsAnswer(&chatter.Answer{})
	it.Then(t).Must(it.Nil(err))

	// Test with actual yield
	jsonData := `{"analysis": "vulnerable to SQL injection", "severity": "high"}`
	answer := &chatter.Answer{
		Yield: []chatter.Json{
			{
				ID:     "security-scan-1",
				Source: "scanner",
				Value:  json.RawMessage(jsonData),
			},
		},
	}

	err = f.AsAnswer(answer)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"text": "regex:.*\"analysis\":\\s*\"vulnerable to SQL injection\".*\"severity\":\\s*\"high\".*"
					}
				]
			}
		]
	}`))
}

func TestEncoderReplyMessage(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	reply := &chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("The code contains a SQL injection vulnerability. Use parameterized queries instead."),
		},
	}

	err = f.AsReply(reply)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"messages": [
			{
				"role": "assistant",
				"content": [
					{
						"text": "The code contains a SQL injection vulnerability. Use parameterized queries instead."
					}
				]
			}
		]
	}`))
}

func TestEncoderConversationFlow(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature: 0.7,
		MaxTokens:   512,
	})

	// System message
	err = f.AsStratum(chatter.Stratum("You are a security expert"))
	it.Then(t).Must(it.Nil(err))

	// User message
	err = f.AsText(chatter.Text("Is this code safe?"))
	it.Then(t).Must(it.Nil(err))

	// Assistant reply
	reply := &chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("I need to see the code to analyze it."),
		},
	}
	err = f.AsReply(reply)
	it.Then(t).Must(it.Nil(err))

	// Another user message
	err = f.AsText(chatter.Text("SELECT * FROM users WHERE id = $1"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"system": [
			{
				"text": "You are a security expert"
			}
		],
		"messages": [
			{
				"role": "user",
				"content": [
					{
						"text": "Is this code safe?"
					}
				]
			},
			{
				"role": "assistant",
				"content": [
					{
						"text": "I need to see the code to analyze it."
					}
				]
			},
			{
				"role": "user",
				"content": [
					{
						"text": "SELECT * FROM users WHERE id = $1"
					}
				]
			}
		],
		"inferenceConfig": {
			"maxTokens": 512,
			"temperature": 0.7
		}
	}`))
}

func TestEncoderUnsupportedFeatures(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	// Test that commands/tools are no-ops for Nova
	f.WithCommand(chatter.Cmd{
		Cmd:    "analyze_code",
		About:  "Analyzes code for vulnerabilities",
		Schema: json.RawMessage(`{"type": "object"}`),
	})

	// Commands should not affect the output
	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"messages": []
	}`))
}
