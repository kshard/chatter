//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text2vec

import (
	"encoding/json"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func TestEncoderBasicTextInput(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Hello world"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-small",
		"input": "Hello world",
		"dimensions": 1024
	}`))
}

func TestEncoderPromptInput(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Summarize the following text")
	prompt.WithInput("Text to summarize:", "The quick brown fox jumps over the lazy dog")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-small",
		"input": "regex:Summarize the following text\\.\\nText to summarize:\\n- The quick brown fox jumps over the lazy dog\\.",
		"dimensions": 1024
	}`))
}

func TestEncoderComplexPromptWithAllContentTypes(t *testing.T) {
	f, err := factory("text-embedding-3-large", 1024)()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Analyze the given code")

	prompt.WithGuide("Guidelines:", "Review for security vulnerabilities", "Check for performance issues")

	prompt.WithRules("Requirements:", "Use secure coding practices", "Follow established patterns")

	prompt.WithExample("func add(a, b int) int { return a + b }", "Simple addition function, no security issues found.")

	prompt.WithContext("Context:", "This is part of a financial application")

	prompt.WithInput("Code to analyze:", "func processPayment(amount float64) error { /* implementation */ }")

	prompt.WithFeedback("Previous feedback:", "Remember to check input validation")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-large",
		"input": "regex:Analyze the given code\\.\\nGuidelines:\\nReview for security vulnerabilities\\.\\nCheck for performance issues\\.\\nRequirements:\\n1\\. Use secure coding practices\\.\\n2\\. Follow established patterns\\.\\nPrevious feedback:\\n- Remember to check input validation\\.\\nExample Input:\\nfunc add\\(a, b int\\) int \\{ return a \\+ b \\}\\nExpected Output:\\nSimple addition function, no security issues found\\.\\n\\n\\nContext:\\n- This is part of a financial application\\.\\nCode to analyze:\\n- func processPayment\\(amount float64\\) error \\{ /\\* implementation \\*/ \\}\\."
	}`))
}

func TestEncoderMultipleTextInputsConcatenation(t *testing.T) {
	f, err := factory("text-embedding-ada-002", 0)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("First part "))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("second part "))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("third part"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-ada-002",
		"input": "First part second part third part"
	}`))
}

func TestEncoderMixedContentTypes(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Initial text. "))
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Process this data")
	prompt.WithInput("Data:", "Sample data content")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text(" Final text."))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-small",
		"input": "regex:Initial text\\. Process this data\\.\\nData:\\n- Sample data content\\. Final text\\."
	}`))
}

func TestEncoderNoOpMethodsValidation(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	// Test that no-op methods don't affect the output
	f.WithInferrer(provider.Inferrer{
		Temperature:   0.7,
		TopP:          0.9,
		TopK:          50,
		MaxTokens:     1000,
		StopSequences: []string{"STOP", "END"},
	})

	f.WithCommand(chatter.Cmd{
		Cmd:    "code_analyzer",
		About:  "Analyzes code for security vulnerabilities",
		Schema: json.RawMessage(`{"type": "object", "properties": {"code": {"type": "string"}}}`),
	})

	// Test all no-op methods
	err = f.AsStratum(chatter.Stratum("You are a helpful AI assistant"))
	it.Then(t).Must(it.Nil(err))

	err = f.AsAnswer(&chatter.Answer{
		Yield: []chatter.Json{
			{
				ID:     "tool-1",
				Source: "calculator",
				Value:  json.RawMessage(`{"result": 42, "operation": "add"}`),
			},
			{
				ID:     "tool-2",
				Source: "validator",
				Value:  json.RawMessage(`{"valid": true, "message": "Input validated"}`),
			},
		},
	})
	it.Then(t).Must(it.Nil(err))

	err = f.AsReply(&chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("The calculation result is 42."),
			chatter.Text(" The input has been validated successfully."),
		},
	})
	it.Then(t).Must(it.Nil(err))

	// Only the actual text content should appear
	err = f.AsText(chatter.Text("This is the only content that should appear in the output"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-small",
		"input": "This is the only content that should appear in the output"
	}`))
}

func TestEncoderEmptyInput(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	// Don't add any content, just build
	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-small",
		"input": ""
	}`))
}

func TestEncoderSpecialCharactersAndEscaping(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Text with \"quotes\" and \n newlines \t tabs & special chars"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-small",
		"input": "regex:Text with \\\"quotes\\\" and \\s+ newlines \\s+ tabs & special chars"
	}`))
}

func TestEncoderUnicodeAndMultilingual(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Hello 世界 🌍 Привет мир مرحبا بالعالم"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-small",
		"input": "Hello 世界 🌍 Привет мир مرحبا بالعالم"
	}`))
}

func TestEncoderLongTextInput(t *testing.T) {
	f, err := factory("text-embedding-3-large", 1024)()
	it.Then(t).Must(it.Nil(err))

	longText := `This is a very long text input that simulates real-world usage scenarios where users might want to embed large documents or extensive content. The purpose is to ensure that the encoder can handle substantial amounts of text without issues. This text contains multiple sentences, various punctuation marks, and represents the kind of content that might be processed in production environments.`

	err = f.AsText(chatter.Text(longText))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"model": "text-embedding-3-large",
		"input": "regex:This is a very long text input that simulates real-world usage scenarios.*production environments\\."
	}`))
}

func TestEncoderDifferentModelTypes(t *testing.T) {
	models := []string{
		"text-embedding-3-small",
		"text-embedding-3-large",
		"text-embedding-ada-002",
		"custom-embedding-model",
	}

	for _, model := range models {
		t.Run("model_"+model, func(t *testing.T) {
			f, err := factory(model, 0)()
			it.Then(t).Must(it.Nil(err))

			err = f.AsText(chatter.Text("Test content for " + model))
			it.Then(t).Must(it.Nil(err))

			result := f.Build()
			it.Then(t).Should(
				it.Equal(result.Model, model),
				it.Equal(result.Text, "Test content for "+model),
			)
		})
	}
}

func TestEncoderBuildIdempotency(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Consistent content"))
	it.Then(t).Must(it.Nil(err))

	// Build multiple times and ensure consistency
	result1 := f.Build()
	result2 := f.Build()
	result3 := f.Build()

	it.Then(t).Should(
		it.Equal(result1.Model, result2.Model),
		it.Equal(result1.Text, result2.Text),
		it.Equal(result2.Model, result3.Model),
		it.Equal(result2.Text, result3.Text),
		it.Equal(result1.Model, "text-embedding-3-small"),
		it.Equal(result1.Text, "Consistent content"),
	)
}

func TestEncoderSequentialOperations(t *testing.T) {
	f, err := factory("text-embedding-3-small", 1024)()
	it.Then(t).Must(it.Nil(err))

	// Test that content accumulates properly
	err = f.AsText(chatter.Text("Part 1"))
	it.Then(t).Must(it.Nil(err))

	intermediateResult := f.Build()
	it.Then(t).Should(
		it.Equal(intermediateResult.Text, "Part 1"),
	)

	err = f.AsText(chatter.Text(" Part 2"))
	it.Then(t).Must(it.Nil(err))

	finalResult := f.Build()
	it.Then(t).Should(
		it.Equal(finalResult.Text, "Part 1 Part 2"),
	)
}

func TestEncoderEdgeCases(t *testing.T) {
	t.Run("empty_string_text", func(t *testing.T) {
		f, err := factory("text-embedding-3-small", 1024)()
		it.Then(t).Must(it.Nil(err))

		err = f.AsText(chatter.Text(""))
		it.Then(t).Must(it.Nil(err))

		it.Then(t).Should(it.Json(f.Build()).Equiv(`{
			"model": "text-embedding-3-small",
			"input": ""
		}`))
	})

	t.Run("whitespace_only_text", func(t *testing.T) {
		f, err := factory("text-embedding-3-small", 1024)()
		it.Then(t).Must(it.Nil(err))

		err = f.AsText(chatter.Text("   \n\t  "))
		it.Then(t).Must(it.Nil(err))

		it.Then(t).Should(it.Json(f.Build()).Equiv(`{
			"model": "text-embedding-3-small",
			"input": "regex:\\s+"
		}`))
	})

	t.Run("empty_model_name", func(t *testing.T) {
		f, err := factory("", 0)()
		it.Then(t).Must(it.Nil(err))

		err = f.AsText(chatter.Text("Test content"))
		it.Then(t).Must(it.Nil(err))

		it.Then(t).Should(it.Json(f.Build()).Equiv(`{
			"model": "",
			"input": "Test content"
		}`))
	})
}
