//
// Copyright 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func TestEncoderInferenceConfiguration(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature:   0.7,
		TopP:          0.9,
		MaxTokens:     1024,
		StopSequences: []string{"stop1", "stop2"},
	})

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(*req.InferenceConfig.Temperature, float32(0.7)),
		it.Equal(*req.InferenceConfig.TopP, float32(0.9)),
		it.Equal(*req.InferenceConfig.MaxTokens, int32(1024)),
		it.Equal(len(req.InferenceConfig.StopSequences), 2),
		it.Equal(req.InferenceConfig.StopSequences[0], "stop1"),
		it.Equal(req.InferenceConfig.StopSequences[1], "stop2"),
		it.Equal(len(req.Messages), 0),
		it.Equal(len(req.System), 0),
	)
}

func TestEncoderSystemMessage(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsStratum(chatter.Stratum("You are a helpful AI assistant"))
	it.Then(t).Must(it.Nil(err))

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.System), 1),
		it.Equal(len(req.Messages), 0),
	)

	// Check system content
	systemBlock := req.System[0].(*types.SystemContentBlockMemberText)
	it.Then(t).Should(it.Equal(systemBlock.Value, "You are a helpful AI assistant"))
}

func TestEncoderUserMessages(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Hello, how can you help me?"))
	it.Then(t).Must(it.Nil(err))

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 1),
		it.Equal(req.Messages[0].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[0].Content), 1),
	)

	// Check content
	contentBlock := req.Messages[0].Content[0].(*types.ContentBlockMemberText)
	it.Then(t).Should(it.Equal(contentBlock.Value, "Hello, how can you help me?"))
}

func TestEncoderPromptMessage(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	var prompt chatter.Prompt
	prompt.WithTask("Analyze the following code for security vulnerabilities")
	prompt.WithInput("Code snippet:", "SELECT * FROM users WHERE id = ?")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 1),
		it.Equal(req.Messages[0].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[0].Content), 1),
	)

	// Check content contains expected text pattern
	contentBlock := req.Messages[0].Content[0].(*types.ContentBlockMemberText)
	promptText := contentBlock.Value
	it.Then(t).Should(
		it.True(len(promptText) > 0),
		it.True(promptText != ""),
	)
}

func TestEncoderAnswerMessage(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	// Test with empty yield (should be no-op)
	err = f.AsAnswer(&chatter.Answer{})
	it.Then(t).Must(it.Nil(err))

	// Test with actual yield
	jsonData := `{"analysis": "vulnerability found", "severity": "high"}`
	answer := &chatter.Answer{
		Yield: []chatter.Json{
			{
				ID:     "tool-call-1",
				Source: "security-scanner",
				Value:  json.RawMessage(jsonData),
			},
		},
	}

	err = f.AsAnswer(answer)
	it.Then(t).Must(it.Nil(err))

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 1),
		it.Equal(req.Messages[0].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[0].Content), 1),
	)

	// Check tool result content
	contentBlock := req.Messages[0].Content[0].(*types.ContentBlockMemberToolResult)
	it.Then(t).Should(
		it.Equal(*contentBlock.Value.ToolUseId, "tool-call-1"),
		it.Equal(len(contentBlock.Value.Content), 1),
	)
}

func TestEncoderReplyMessage(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	reply := &chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("The code is vulnerable to SQL injection. Use parameterized queries."),
		},
	}

	err = f.AsReply(reply)
	it.Then(t).Must(it.Nil(err))

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 1),
		it.Equal(req.Messages[0].Role, types.ConversationRoleAssistant),
		it.Equal(len(req.Messages[0].Content), 1),
	)

	// Check content
	contentBlock := req.Messages[0].Content[0].(*types.ContentBlockMemberText)
	it.Then(t).Should(it.Equal(contentBlock.Value, "The code is vulnerable to SQL injection. Use parameterized queries."))
}

func TestEncoderToolConfiguration(t *testing.T) {
	registry := chatter.Registry{
		{
			Cmd:    "code_analyzer",
			About:  "Analyzes code for security vulnerabilities",
			Schema: json.RawMessage(`{"type": "object", "properties": {"code": {"type": "string"}}}`),
		},
	}

	f, err := factory("test-model", registry)()
	it.Then(t).Must(it.Nil(err))

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 0),
		it.Equal(len(req.System), 0),
		it.True(req.ToolConfig != nil),
		it.Equal(len(req.ToolConfig.Tools), 1),
	)

	// Check tool configuration
	tool := req.ToolConfig.Tools[0].(*types.ToolMemberToolSpec)
	it.Then(t).Should(
		it.Equal(*tool.Value.Name, "code_analyzer"),
		it.Equal(*tool.Value.Description, "Analyzes code for security vulnerabilities"),
	)
}

func TestEncoderWithCommand(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	f.WithCommand(chatter.Cmd{
		Cmd:    "weather_check",
		About:  "Checks weather information for a location",
		Schema: json.RawMessage(`{"type": "object", "properties": {"location": {"type": "string"}}}`),
	})

	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 0),
		it.Equal(len(req.System), 0),
		it.True(req.ToolConfig != nil),
		it.Equal(len(req.ToolConfig.Tools), 1),
	)

	// Check tool configuration
	tool := req.ToolConfig.Tools[0].(*types.ToolMemberToolSpec)
	it.Then(t).Should(
		it.Equal(*tool.Value.Name, "weather_check"),
		it.Equal(*tool.Value.Description, "Checks weather information for a location"),
	)
}

func TestEncoderConversationFlow(t *testing.T) {
	f, err := factory("anthropic.claude-3-sonnet", nil)()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature: 0.5,
		MaxTokens:   512,
	})

	// System message
	err = f.AsStratum(chatter.Stratum("You are a helpful code review assistant"))
	it.Then(t).Must(it.Nil(err))

	// User message
	err = f.AsText(chatter.Text("Can you review this function?"))
	it.Then(t).Must(it.Nil(err))

	// Assistant reply
	reply := &chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("Sure! Please share the function code."),
		},
	}
	err = f.AsReply(reply)
	it.Then(t).Must(it.Nil(err))

	// Another user message with prompt
	var prompt chatter.Prompt
	prompt.WithTask("Review this function")
	prompt.WithInput("Function:", "func add(a, b int) int { return a + b }")

	err = f.AsPrompt(&prompt)
	it.Then(t).Must(it.Nil(err))

	// Build and validate the conversation flow
	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "anthropic.claude-3-sonnet"),
		it.Equal(*req.InferenceConfig.Temperature, float32(0.5)),
		it.Equal(*req.InferenceConfig.MaxTokens, int32(512)),
		it.Equal(len(req.System), 1),
		it.Equal(len(req.Messages), 3),
	)

	// Validate system message
	systemContent := req.System[0].(*types.SystemContentBlockMemberText)
	it.Then(t).Should(it.Equal(systemContent.Value, "You are a helpful code review assistant"))

	// Validate first user message
	it.Then(t).Should(
		it.Equal(req.Messages[0].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[0].Content), 1),
	)
	userContent1 := req.Messages[0].Content[0].(*types.ContentBlockMemberText)
	it.Then(t).Should(it.Equal(userContent1.Value, "Can you review this function?"))

	// Validate assistant reply
	it.Then(t).Should(
		it.Equal(req.Messages[1].Role, types.ConversationRoleAssistant),
		it.Equal(len(req.Messages[1].Content), 1),
	)
	assistantContent := req.Messages[1].Content[0].(*types.ContentBlockMemberText)
	it.Then(t).Should(it.Equal(assistantContent.Value, "Sure! Please share the function code."))

	// Validate second user message (prompt)
	it.Then(t).Should(
		it.Equal(req.Messages[2].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[2].Content), 1),
	)
	userContent2 := req.Messages[2].Content[0].(*types.ContentBlockMemberText)
	it.Then(t).Should(
		it.True(strings.Contains(userContent2.Value, "Review this function")),
		it.True(strings.Contains(userContent2.Value, "func add(a, b int) int")),
	)
}

func TestEncoderComplexToolWorkflow(t *testing.T) {
	registry := chatter.Registry{
		{
			Cmd:    "file_reader",
			About:  "Reads the contents of a file",
			Schema: json.RawMessage(`{"type": "object", "properties": {"filename": {"type": "string"}}}`),
		},
	}

	f, err := factory("test-model", registry)()
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{
		Temperature: 0.3,
		TopP:        0.8,
	})

	// Add additional tool via WithCommand
	f.WithCommand(chatter.Cmd{
		Cmd:    "syntax_checker",
		About:  "Checks syntax of code",
		Schema: json.RawMessage(`{"type": "object", "properties": {"language": {"type": "string"}, "code": {"type": "string"}}}`),
	})

	// User request
	err = f.AsText(chatter.Text("Read the file main.go and check its syntax"))
	it.Then(t).Must(it.Nil(err))

	// Tool response
	toolResponse := `{"content": "package main\nfunc main() { fmt.Println(\"Hello\") }"}`
	answer := &chatter.Answer{
		Yield: []chatter.Json{
			{
				ID:     "file-read-1",
				Source: "file_reader",
				Value:  json.RawMessage(toolResponse),
			},
		},
	}
	err = f.AsAnswer(answer)
	it.Then(t).Must(it.Nil(err))

	// Build and validate the complex tool workflow
	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(*req.InferenceConfig.Temperature, float32(0.3)),
		it.Equal(*req.InferenceConfig.TopP, float32(0.8)),
		it.Equal(len(req.Messages), 2),
	)

	// Validate tool configuration
	it.Then(t).Should(
		it.True(req.ToolConfig != nil),
		it.Equal(len(req.ToolConfig.Tools), 2),
	)

	// Validate first tool (file_reader)
	tool1 := req.ToolConfig.Tools[0].(*types.ToolMemberToolSpec)
	it.Then(t).Should(
		it.Equal(*tool1.Value.Name, "file_reader"),
		it.Equal(*tool1.Value.Description, "Reads the contents of a file"),
	)

	// Validate second tool (syntax_checker)
	tool2 := req.ToolConfig.Tools[1].(*types.ToolMemberToolSpec)
	it.Then(t).Should(
		it.Equal(*tool2.Value.Name, "syntax_checker"),
		it.Equal(*tool2.Value.Description, "Checks syntax of code"),
	)

	// Validate first user message
	it.Then(t).Should(
		it.Equal(req.Messages[0].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[0].Content), 1),
	)
	userContent := req.Messages[0].Content[0].(*types.ContentBlockMemberText)
	it.Then(t).Should(it.Equal(userContent.Value, "Read the file main.go and check its syntax"))

	// Validate tool result message
	it.Then(t).Should(
		it.Equal(req.Messages[1].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[1].Content), 1),
	)
	toolResult := req.Messages[1].Content[0].(*types.ContentBlockMemberToolResult)
	it.Then(t).Should(
		it.Equal(*toolResult.Value.ToolUseId, "file-read-1"),
		it.Equal(len(toolResult.Value.Content), 1),
	)
}

func TestEncoderInvalidToolSchema(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	// WithCommand should handle invalid JSON gracefully by logging and continuing
	f.WithCommand(chatter.Cmd{
		Cmd:    "invalid_tool",
		About:  "Tool with invalid JSON schema",
		Schema: json.RawMessage(`{invalid json}`),
	})

	// Should still build successfully, but without the invalid tool
	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 0),
		it.Equal(len(req.System), 0),
	)
	// Tool config should be nil or empty since the invalid tool was rejected
	it.Then(t).Should(it.True(req.ToolConfig == nil || len(req.ToolConfig.Tools) == 0))
}

func TestEncoderMultipleAnswersInSingleMessage(t *testing.T) {
	f, err := factory("test-model", nil)()
	it.Then(t).Must(it.Nil(err))

	answer := &chatter.Answer{
		Yield: []chatter.Json{
			{
				ID:     "tool-1",
				Source: "analyzer",
				Value:  json.RawMessage(`{"result": "pass"}`),
			},
			{
				ID:     "tool-2",
				Source: "validator",
				Value:  json.RawMessage(`{"valid": true}`),
			},
		},
	}

	err = f.AsAnswer(answer)
	it.Then(t).Must(it.Nil(err))

	// Build and validate multiple answers in single message
	req := f.Build()
	it.Then(t).Should(
		it.Equal(*req.ModelId, "test-model"),
		it.Equal(len(req.Messages), 1),
	)

	// Validate the message has multiple tool results
	it.Then(t).Should(
		it.Equal(req.Messages[0].Role, types.ConversationRoleUser),
		it.Equal(len(req.Messages[0].Content), 2),
	)

	// Validate first tool result
	toolResult1 := req.Messages[0].Content[0].(*types.ContentBlockMemberToolResult)
	it.Then(t).Should(
		it.Equal(*toolResult1.Value.ToolUseId, "tool-1"),
		it.Equal(len(toolResult1.Value.Content), 1),
	)

	// Validate second tool result
	toolResult2 := req.Messages[0].Content[1].(*types.ContentBlockMemberToolResult)
	it.Then(t).Should(
		it.Equal(*toolResult2.Value.ToolUseId, "tool-2"),
		it.Equal(len(toolResult2.Value.Content), 1),
	)
}
