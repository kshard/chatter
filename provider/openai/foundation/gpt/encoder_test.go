//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package gpt

import (
	"encoding/json"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func TestEncoderBasicConfiguration(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": []
	}`))
}

func TestEncoderInferenceConfiguration(t *testing.T) {
	f, err := factory("gpt-3.5-turbo")()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature: 0.7,
		TopP:        0.9,
		MaxTokens:   512,
	})

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-3.5-turbo",
		"messages": [],
		"temperature": 0.7,
		"top_p": 0.9,
		"max_tokens": 512
	}`))
}

func TestEncoderSystemMessage(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	err = f.AsStratum(chatter.Stratum("You are a helpful AI assistant specialized in code analysis"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "system",
				"content": "You are a helpful AI assistant specialized in code analysis"
			}
		]
	}`))
}

func TestEncoderUserMessages(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Explain quantum computing in simple terms"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "user",
				"content": "Explain quantum computing in simple terms"
			}
		]
	}`))
}

func TestEncoderPromptMessage(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Analyze the following code for security vulnerabilities")
	prompt.WithInput("Code snippet:", "SELECT * FROM users WHERE id = ?")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "user",
				"content": "regex:Analyze the following code for security vulnerabilities\\.\\nCode snippet:\\n- SELECT \\* FROM users WHERE id = \\?"
			}
		]
	}`))
}

func TestEncoderComplexPromptWithAllContentTypes(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Review the following code")

	prompt.WithGuide("Guidelines:", "Check for security issues", "Verify error handling")

	prompt.WithRules("Requirements:", "Use secure coding practices", "Follow best practices")

	prompt.WithExample("func add(a, b int) int { return a + b }", "Simple addition function, no issues found.")

	prompt.WithContext("Context:", "This is part of a financial application")

	prompt.WithInput("Code to review:", "func processPayment(amount float64) error { /* implementation */ }")

	prompt.WithFeedback("Previous feedback:", "Remember to validate inputs")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "user",
				"content": "regex:Review the following code\\.\\nGuidelines:\\nCheck for security issues\\.\\nVerify error handling\\.\\nRequirements:\\n1\\. Use secure coding practices\\.\\n2\\. Follow best practices\\.\\nPrevious feedback:\\n- Remember to validate inputs\\.\\nExample Input:\\nfunc add\\(a, b int\\) int \\{ return a \\+ b \\}\\nExpected Output:\\nSimple addition function, no issues found\\.\\n\\n\\nContext:\\n- This is part of a financial application\\.\\nCode to review:\\n- func processPayment\\(amount float64\\) error \\{ /\\* implementation \\*/ \\}\\."
			}
		]
	}`))
}

func TestEncoderReplyMessage(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	reply := &chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("The function looks good. No security issues found."),
		},
	}

	err = f.AsReply(reply)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "assistant",
				"content": "The function looks good. No security issues found."
			}
		]
	}`))
}

func TestEncoderConversationFlow(t *testing.T) {
	f, err := factory("gpt-3.5-turbo")()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature: 0.5,
		MaxTokens:   256,
	})

	err = f.AsStratum(chatter.Stratum("You are a helpful code review assistant"))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Can you review this function?"))
	it.Then(t).Must(it.Nil(err))

	err = f.AsReply(&chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("Sure! Please share the function code."),
		},
	})
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Review this function")
	prompt.WithInput("Function:", "func add(a, b int) int { return a + b }")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	result := f.Build()
	it.Then(t).Should(
		it.Equal(result.Model, "gpt-3.5-turbo"),
		it.Equal(result.Temperature, 0.5),
		it.Equal(result.MaxTokens, 256),
		it.Equal(len(result.Messages), 4),
	)

	// Check individual message types and roles
	it.Then(t).Should(
		it.Equal(result.Messages[0].Role, "system"),
		it.Equal(result.Messages[0].Content, "You are a helpful code review assistant"),
		it.Equal(result.Messages[1].Role, "user"),
		it.Equal(result.Messages[1].Content, "Can you review this function?"),
		it.Equal(result.Messages[2].Role, "assistant"),
		it.Equal(result.Messages[2].Content, "Sure! Please share the function code."),
		it.Equal(result.Messages[3].Role, "user"),
	)

	// Verify the prompt content pattern
	it.Then(t).Should(
		it.True(len(result.Messages[3].Content) > 0),
	)
}

func TestEncoderNoOpMethods(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	// Test that WithCommand is a no-op (not supported yet)
	f.WithCommand(chatter.Cmd{
		Cmd:    "code_analyzer",
		About:  "Analyzes code for security vulnerabilities",
		Schema: json.RawMessage(`{"type": "object", "properties": {"code": {"type": "string"}}}`),
	})

	// Test that AsAnswer is a no-op (not supported yet)
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

	// Add some actual content
	err = f.AsText(chatter.Text("Hello world"))
	it.Then(t).Must(it.Nil(err))

	// Only the text message should appear, no command or answer effects
	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "user",
				"content": "Hello world"
			}
		]
	}`))
}

func TestEncoderDifferentModelTypes(t *testing.T) {
	models := []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"gpt-4o",
		"custom-model",
	}

	for _, model := range models {
		t.Run("model_"+model, func(t *testing.T) {
			f, err := factory(model)()
			it.Then(t).Must(it.Nil(err))

			err = f.AsText(chatter.Text("Test message for " + model))
			it.Then(t).Must(it.Nil(err))

			result := f.Build()
			it.Then(t).Should(
				it.Equal(result.Model, model),
				it.Equal(len(result.Messages), 1),
				it.Equal(result.Messages[0].Role, "user"),
				it.Equal(result.Messages[0].Content, "Test message for "+model),
			)
		})
	}
}

func TestEncoderSpecialCharacters(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Text with \"quotes\" and \n newlines \t tabs & special chars"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "user",
				"content": "regex:Text with \\\"quotes\\\" and \\s+ newlines \\s+ tabs & special chars"
			}
		]
	}`))
}

func TestEncoderMultipleMessagesSequence(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	// Add multiple messages in sequence
	err = f.AsStratum(chatter.Stratum("You are an expert programmer"))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("What is a function?"))
	it.Then(t).Must(it.Nil(err))

	err = f.AsReply(&chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("A function is a reusable block of code."),
		},
	})
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Can you give an example?"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "gpt-4",
		"messages": [
			{
				"role": "system",
				"content": "You are an expert programmer"
			},
			{
				"role": "user",
				"content": "What is a function?"
			},
			{
				"role": "assistant",
				"content": "A function is a reusable block of code."
			},
			{
				"role": "user",
				"content": "Can you give an example?"
			}
		]
	}`))
}

func TestEncoderBuildIdempotency(t *testing.T) {
	f, err := factory("gpt-4")()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Consistent message"))
	it.Then(t).Must(it.Nil(err))

	// Build multiple times and ensure consistency
	result1 := f.Build()
	result2 := f.Build()
	result3 := f.Build()

	it.Then(t).Should(
		it.Equal(result1.Model, result2.Model),
		it.Equal(len(result1.Messages), len(result2.Messages)),
		it.Equal(result2.Model, result3.Model),
		it.Equal(len(result2.Messages), len(result3.Messages)),
		it.Equal(result1.Model, "gpt-4"),
		it.Equal(len(result1.Messages), 1),
		it.Equal(result1.Messages[0].Content, "Consistent message"),
	)
}

func TestEncoderEdgeCases(t *testing.T) {
	t.Run("empty_model_name", func(t *testing.T) {
		f, err := factory("")()
		it.Then(t).Must(it.Nil(err))

		err = f.AsText(chatter.Text("Test content"))
		it.Then(t).Must(it.Nil(err))

		it.Then(t).Should(it.Json(f.Build()).Equiv(`{
			"model": "",
			"messages": [
				{
					"role": "user",
					"content": "Test content"
				}
			]
		}`))
	})

	t.Run("empty_text_message", func(t *testing.T) {
		f, err := factory("gpt-4")()
		it.Then(t).Must(it.Nil(err))

		err = f.AsText(chatter.Text(""))
		it.Then(t).Must(it.Nil(err))

		it.Then(t).Should(it.Json(f.Build()).Equiv(`{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": ""
				}
			]
		}`))
	})

	t.Run("whitespace_only_message", func(t *testing.T) {
		f, err := factory("gpt-4")()
		it.Then(t).Must(it.Nil(err))

		err = f.AsText(chatter.Text("   \n\t  "))
		it.Then(t).Must(it.Nil(err))

		it.Then(t).Should(it.Json(f.Build()).Equiv(`{
			"model": "gpt-4",
			"messages": [
				{
					"role": "user",
					"content": "regex:\\s+"
				}
			]
		}`))
	})
}
