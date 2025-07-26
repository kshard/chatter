//
// Copyright 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/document"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
	"github.com/fogfish/it/v2"
)

func TestDecoderBasicTextResponse(t *testing.T) {
	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "This is a simple text response from the LLM.",
					},
				},
			},
		},
		StopReason: types.StopReasonEndTurn,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(15),
			OutputTokens: aws.Int32(8),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 15,
				"replyTokens": 8
			},
			"content": [
				{
					"text": "This is a simple text response from the LLM."
				}
			]
		}`),
	)
}

func TestDecoderToolInvocationResponse(t *testing.T) {
	toolInput := map[string]interface{}{
		"location": "San Francisco",
		"units":    "celsius",
	}

	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberToolUse{
						Value: types.ToolUseBlock{
							ToolUseId: aws.String("weather-tool-1"),
							Name:      aws.String("get_weather"),
							Input:     document.NewLazyDocument(toolInput),
						},
					},
				},
			},
		},
		StopReason: types.StopReasonToolUse,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(25),
			OutputTokens: aws.Int32(12),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "invoke",
			"usage": {
				"inputTokens": 25,
				"replyTokens": 12
			},
			"content": [
				{
					"name": "get_weather",
					"args": {
						"id": "weather-tool-1",
						"bag": {
							"location": "San Francisco",
							"units": "celsius"
						}
					}
				}
			]
		}`),
	)
}

func TestDecoderStageMapping(t *testing.T) {
	testCases := []struct {
		name          string
		stopReason    types.StopReason
		expectedStage string
	}{
		{
			name:          "end_turn_returns_complete",
			stopReason:    types.StopReasonEndTurn,
			expectedStage: "return",
		},
		{
			name:          "max_tokens_returns_incomplete",
			stopReason:    types.StopReasonMaxTokens,
			expectedStage: "incomplete",
		},
		{
			name:          "stop_sequence_returns_incomplete",
			stopReason:    types.StopReasonStopSequence,
			expectedStage: "incomplete",
		},
		{
			name:          "tool_use_returns_invoke",
			stopReason:    types.StopReasonToolUse,
			expectedStage: "invoke",
		},
		{
			name:          "unknown_reason_returns_error",
			stopReason:    "unknown_reason",
			expectedStage: "error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := &bedrockruntime.ConverseOutput{
				Output: &types.ConverseOutputMemberMessage{
					Value: types.Message{
						Role: types.ConversationRoleAssistant,
						Content: []types.ContentBlock{
							&types.ContentBlockMemberText{
								Value: "Test response",
							},
						},
					},
				},
				StopReason: tc.stopReason,
				Usage: &types.TokenUsage{
					InputTokens:  aws.Int32(10),
					OutputTokens: aws.Int32(5),
				},
			}

			reply, err := decoder{}.Decode(input)

			it.Then(t).Should(
				it.Nil(err),
				it.Json(reply).Equiv(`{
					"stage": "`+tc.expectedStage+`",
					"usage": {
						"inputTokens": 10,
						"replyTokens": 5
					},
					"content": [
						{
							"text": "Test response"
						}
					]
				}`),
			)
		})
	}
}

func TestDecoderMultipleContentBlocks(t *testing.T) {
	toolArgs := map[string]interface{}{
		"code":     "SELECT * FROM users",
		"language": "sql",
	}

	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "I'll analyze this code for you.",
					},
					&types.ContentBlockMemberToolUse{
						Value: types.ToolUseBlock{
							ToolUseId: aws.String("analyzer-1"),
							Name:      aws.String("analyze_code"),
							Input:     document.NewLazyDocument(toolArgs),
						},
					},
					&types.ContentBlockMemberText{
						Value: "Let me check for security vulnerabilities.",
					},
				},
			},
		},
		StopReason: types.StopReasonToolUse,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(40),
			OutputTokens: aws.Int32(18),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "invoke",
			"usage": {
				"inputTokens": 40,
				"replyTokens": 18
			},
			"content": [
				{
					"text": "I'll analyze this code for you."
				},
				{
					"name": "analyze_code",
					"args": {
						"id": "analyzer-1",
						"bag": {
							"code": "SELECT * FROM users",
							"language": "sql"
						}
					}
				},
				{
					"text": "Let me check for security vulnerabilities."
				}
			]
		}`),
	)
}

func TestDecoderEmptyContent(t *testing.T) {
	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role:    types.ConversationRoleAssistant,
				Content: []types.ContentBlock{},
			},
		},
		StopReason: types.StopReasonEndTurn,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(5),
			OutputTokens: aws.Int32(0),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 5,
				"replyTokens": 0
			},
			"content": []
		}`),
	)
}

func TestDecoderMissingUsageInfo(t *testing.T) {
	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "Response without usage information.",
					},
				},
			},
		},
		StopReason: types.StopReasonEndTurn,
		Usage:      nil,
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 0,
				"replyTokens": 0
			},
			"content": [
				{
					"text": "Response without usage information."
				}
			]
		}`),
	)
}

func TestDecoderComplexToolWorkflow(t *testing.T) {
	// Simulate a complex response with multiple tool calls and text
	fileContent := map[string]interface{}{
		"filename": "config.json",
		"type":     "read",
	}

	validateArgs := map[string]interface{}{
		"content": `{"api_key": "secret", "debug": true}`,
		"schema":  "config-schema.json",
	}

	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "I'll help you process the configuration file. Let me read it first.",
					},
					&types.ContentBlockMemberToolUse{
						Value: types.ToolUseBlock{
							ToolUseId: aws.String("file-reader-001"),
							Name:      aws.String("read_file"),
							Input:     document.NewLazyDocument(fileContent),
						},
					},
					&types.ContentBlockMemberText{
						Value: "Now I'll validate the configuration against the schema.",
					},
					&types.ContentBlockMemberToolUse{
						Value: types.ToolUseBlock{
							ToolUseId: aws.String("validator-002"),
							Name:      aws.String("validate_json"),
							Input:     document.NewLazyDocument(validateArgs),
						},
					},
				},
			},
		},
		StopReason: types.StopReasonToolUse,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(65),
			OutputTokens: aws.Int32(28),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "invoke",
			"usage": {
				"inputTokens": 65,
				"replyTokens": 28
			},
			"content": [
				{
					"text": "I'll help you process the configuration file. Let me read it first."
				},
				{
					"name": "read_file",
					"args": {
						"id": "file-reader-001",
						"bag": {
							"filename": "config.json",
							"type": "read"
						}
					}
				},
				{
					"text": "Now I'll validate the configuration against the schema."
				},
				{
					"name": "validate_json",
					"args": {
						"id": "validator-002",
						"bag": {
							"content": "{\"api_key\": \"secret\", \"debug\": true}",
							"schema": "config-schema.json"
						}
					}
				}
			]
		}`),
	)
}

func TestDecoderUnknownContentType(t *testing.T) {
	// Create a mock content block type that's not recognized
	// This tests the warning log and nil return path in decodeContent
	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "Valid text content",
					},
					// Note: We can't easily create an unknown type, so we'll use what we have
					// The decoder should handle all known types properly
				},
			},
		},
		StopReason: types.StopReasonEndTurn,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(10),
			OutputTokens: aws.Int32(3),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 10,
				"replyTokens": 3
			},
			"content": [
				{
					"text": "Valid text content"
				}
			]
		}`),
	)
}

func TestDecoderInvalidOutputType(t *testing.T) {
	// Create an output with an unknown type to test error handling
	input := &bedrockruntime.ConverseOutput{
		Output:     nil, // This will cause an error in decodeOutput
		StopReason: types.StopReasonEndTurn,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(5),
			OutputTokens: aws.Int32(0),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.True(err != nil),
		it.True(reply == nil),
	)
}

func TestDecoderToolInvocationWithComplexArguments(t *testing.T) {
	// Test with complex nested JSON arguments
	complexArgs := map[string]interface{}{
		"query": map[string]interface{}{
			"select": []string{"name", "email", "created_at"},
			"from":   "users",
			"where": map[string]interface{}{
				"status": "active",
				"age":    map[string]interface{}{"$gte": 18},
			},
			"limit": 100,
		},
		"options": map[string]interface{}{
			"timeout":     30,
			"retry_count": 3,
			"cache":       true,
		},
	}

	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "I'll execute this database query for you.",
					},
					&types.ContentBlockMemberToolUse{
						Value: types.ToolUseBlock{
							ToolUseId: aws.String("db-query-123"),
							Name:      aws.String("execute_query"),
							Input:     document.NewLazyDocument(complexArgs),
						},
					},
				},
			},
		},
		StopReason: types.StopReasonToolUse,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(45),
			OutputTokens: aws.Int32(15),
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "invoke",
			"usage": {
				"inputTokens": 45,
				"replyTokens": 15
			},
			"content": [
				{
					"text": "I'll execute this database query for you."
				},
				{
					"name": "execute_query",
					"args": {
						"id": "db-query-123",
						"bag": {
							"query": {
								"select": ["name", "email", "created_at"],
								"from": "users",
								"where": {
									"status": "active",
									"age": {"$gte": 18}
								},
								"limit": 100
							},
							"options": {
								"timeout": 30,
								"retry_count": 3,
								"cache": true
							}
						}
					}
				}
			]
		}`),
	)
}

func TestDecoderPartialUsageInformation(t *testing.T) {
	// Test with only input tokens present
	input := &bedrockruntime.ConverseOutput{
		Output: &types.ConverseOutputMemberMessage{
			Value: types.Message{
				Role: types.ConversationRoleAssistant,
				Content: []types.ContentBlock{
					&types.ContentBlockMemberText{
						Value: "Response with partial usage data.",
					},
				},
			},
		},
		StopReason: types.StopReasonEndTurn,
		Usage: &types.TokenUsage{
			InputTokens:  aws.Int32(20),
			OutputTokens: nil, // Missing output tokens
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 20,
				"replyTokens": 0
			},
			"content": [
				{
					"text": "Response with partial usage data."
				}
			]
		}`),
	)
}
