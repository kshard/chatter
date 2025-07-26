//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package nova

import (
	"testing"

	"github.com/fogfish/it/v2"
)

func TestDecoderSuccessfulCompletion(t *testing.T) {
	input := &reply{
		Output: output{
			Message: message{
				Role: "assistant",
				Content: []content{
					{Text: "The capital of France is Paris."},
				},
			},
		},
		Usage: usage{
			InputTokens:  15,
			OutputTokens: 8,
			TotalTokens:  23,
		},
		StopReason: "end_turn",
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
				{"text": "The capital of France is Paris."}
			]
		}`),
	)
}

func TestDecoderIncompleteResponses(t *testing.T) {
	t.Run("max_tokens_exceeded", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: "The capital of France is"},
					},
				},
			},
			Usage: usage{
				InputTokens:  10,
				OutputTokens: 5,
			},
			StopReason: "max_tokens",
		}

		reply, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(reply).Equiv(`{
				"stage": "incomplete",
				"usage": {
					"inputTokens": 10,
					"replyTokens": 5
				},
				"content": [
					{"text": "The capital of France is"}
				]
			}`),
		)
	})

	t.Run("stop_sequence_triggered", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: "Here is the answer [STOP]"},
					},
				},
			},
			Usage: usage{
				InputTokens:  12,
				OutputTokens: 6,
			},
			StopReason: "stop_sequence",
		}

		reply, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(reply).Equiv(`{
				"stage": "incomplete",
				"usage": {
					"inputTokens": 12,
					"replyTokens": 6
				},
				"content": [
					{"text": "Here is the answer [STOP]"}
				]
			}`),
		)
	})
}

func TestDecoderErrorConditions(t *testing.T) {
	t.Run("unknown_stop_reason", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: "Partial response"},
					},
				},
			},
			Usage: usage{
				InputTokens:  8,
				OutputTokens: 2,
			},
			StopReason: "unknown_error",
		}

		reply, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(reply).Equiv(`{
				"stage": "error",
				"usage": {
					"inputTokens": 8,
					"replyTokens": 2
				},
				"content": [
					{"text": "Partial response"}
				]
			}`),
		)
	})

	t.Run("content_processing_error", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: ""},
					},
				},
			},
			Usage: usage{
				InputTokens:  5,
				OutputTokens: 0,
			},
			StopReason: "server_error",
		}

		reply, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(reply).Equiv(`{
				"stage": "error",
				"usage": {
					"inputTokens": 5,
					"replyTokens": 0
				},
				"content": []
			}`),
		)
	})
}

func TestDecoderContentProcessing(t *testing.T) {
	t.Run("multiple_text_blocks", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: "First part of the response. "},
						{Text: "Second part of the response."},
					},
				},
			},
			Usage: usage{
				InputTokens:  20,
				OutputTokens: 12,
			},
			StopReason: "end_turn",
		}

		reply, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(reply).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 20,
					"replyTokens": 12
				},
				"content": [
					{"text": "First part of the response. "},
					{"text": "Second part of the response."}
				]
			}`),
		)
	})

	t.Run("mixed_content_with_empty_text", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: "Valid text content"},
						{Text: ""},
						{Text: "Another valid text"},
						{Image: map[string]interface{}{"format": "jpeg"}},
					},
				},
			},
			Usage: usage{
				InputTokens:  15,
				OutputTokens: 8,
			},
			StopReason: "end_turn",
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
					{"text": "Valid text content"},
					{"text": "Another valid text"}
				]
			}`),
		)
	})

	t.Run("no_valid_text_content", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: ""},
						{Image: map[string]interface{}{"format": "png"}},
						{Video: map[string]interface{}{"format": "mp4"}},
					},
				},
			},
			Usage: usage{
				InputTokens:  10,
				OutputTokens: 0,
			},
			StopReason: "end_turn",
		}

		reply, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(reply).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 10,
					"replyTokens": 0
				},
				"content": []
			}`),
		)
	})
}

func TestDecoderTokenUsageScenarios(t *testing.T) {
	t.Run("high_token_usage", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role: "assistant",
					Content: []content{
						{Text: "This is a comprehensive response that uses many tokens to provide detailed information about the requested topic, including examples and explanations."},
					},
				},
			},
			Usage: usage{
				InputTokens:  150,
				OutputTokens: 75,
				TotalTokens:  225,
			},
			StopReason: "end_turn",
		}

		reply, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(reply).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 150,
					"replyTokens": 75
				},
				"content": [
					{"text": "regex:This is a comprehensive response.*detailed information.*examples and explanations\\."}
				]
			}`),
		)
	})

	t.Run("zero_token_response", func(t *testing.T) {
		input := &reply{
			Output: output{
				Message: message{
					Role:    "assistant",
					Content: []content{},
				},
			},
			Usage: usage{
				InputTokens:  5,
				OutputTokens: 0,
			},
			StopReason: "end_turn",
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
	})
}

func TestDecoderStageMapping(t *testing.T) {
	testCases := []struct {
		name       string
		stopReason string
		expected   string
	}{
		{"successful_completion", "end_turn", "return"},
		{"max_tokens_limit", "max_tokens", "incomplete"},
		{"stop_sequence_hit", "stop_sequence", "incomplete"},
		{"service_timeout", "timeout", "error"},
		{"api_error", "api_error", "error"},
		{"empty_stop_reason", "", "error"},
		{"null_stop_reason", "null", "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := &reply{
				Output: output{
					Message: message{
						Role: "assistant",
						Content: []content{
							{Text: "Test response"},
						},
					},
				},
				Usage: usage{
					InputTokens:  10,
					OutputTokens: 3,
				},
				StopReason: tc.stopReason,
			}

			reply, err := decoder{}.Decode(input)

			it.Then(t).Should(
				it.Nil(err),
				it.Json(reply).Equiv(`{
					"stage": "`+tc.expected+`",
					"usage": {
						"inputTokens": 10,
						"replyTokens": 3
					},
					"content": [
						{"text": "Test response"}
					]
				}`),
			)
		})
	}
}
