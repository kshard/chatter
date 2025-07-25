package titan

import (
	"encoding/json"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func TestEncoderTextInput(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Hello world"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "Hello world"
	}`))
}

func TestEncoderPromptInput(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Summarize the following text")
	prompt.WithInput("Text to summarize:", "The quick brown fox jumps over the lazy dog")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "regex:Summarize the following text.\nText to summarize:\n- The quick brown fox jumps over the lazy dog."
	}`))
}

func TestEncoderMultipleTextInputs(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("First part "))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("second part "))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("third part"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "First part second part third part"
	}`))
}

func TestEncoderNoOpMethods(t *testing.T) {
	f, err := factory()
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
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	// Don't add any content, just build
	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": ""
	}`))
}

func TestEncoderSpecialCharacters(t *testing.T) {
	f, err := factory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Text with \"quotes\" and \n newlines \t tabs"))
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(it.Json(f.Build()).Equiv(`{
		"inputText": "regex:Text with \\\"quotes\\\" and \\s+ newlines \\s+ tabs"
	}`))
}
