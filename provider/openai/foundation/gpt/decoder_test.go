//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package gpt

import (
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
)

func TestDecoderBasicReplyDecoding(t *testing.T) {
	input := &reply{
		ID: "chatcmpl-test-123",
		Choices: []choice{
			{
				Message: message{
					Role:    "assistant",
					Content: "Hello! I'm an AI assistant. How can I help you today?",
				},
			},
		},
		Usage: usage{
			PromptTokens: 15,
			OutputTokens: 12,
			UsedTokens:   27,
		},
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 15,
				"replyTokens": 12
			},
			"content": [
				{
					"text": "Hello! I'm an AI assistant. How can I help you today?"
				}
			]
		}`),
	)
}

func TestDecoderComplexResponseContent(t *testing.T) {
	input := &reply{
		ID: "chatcmpl-code-review-456",
		Choices: []choice{
			{
				Message: message{
					Role:    "assistant",
					Content: `Based on my analysis of the code.`,
				},
			},
		},
		Usage: usage{
			PromptTokens: 250,
			OutputTokens: 75,
			UsedTokens:   325,
		},
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 250,
				"replyTokens": 75
			},
			"content": [
				{
					"text": "Based on my analysis of the code."
				}
			]
		}`),
	)
}

func TestDecoderUsageTokenMapping(t *testing.T) {
	tests := []struct {
		name         string
		promptTokens int
		outputTokens int
		usedTokens   int
	}{
		{
			name:         "standard_tokens",
			promptTokens: 100,
			outputTokens: 50,
			usedTokens:   150,
		},
		{
			name:         "zero_output_tokens",
			promptTokens: 25,
			outputTokens: 0,
			usedTokens:   25,
		},
		{
			name:         "high_token_usage",
			promptTokens: 4000,
			outputTokens: 1000,
			usedTokens:   5000,
		},
		{
			name:         "minimal_tokens",
			promptTokens: 1,
			outputTokens: 1,
			usedTokens:   2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			input := &reply{
				ID: "chatcmpl-token-test",
				Choices: []choice{
					{
						Message: message{
							Role:    "assistant",
							Content: "Token usage test response",
						},
					},
				},
				Usage: usage{
					PromptTokens: test.promptTokens,
					OutputTokens: test.outputTokens,
					UsedTokens:   test.usedTokens,
				},
			}

			result, err := decoder{}.Decode(input)

			it.Then(t).Should(
				it.Nil(err),
				it.Equal(result.Usage.InputTokens, test.promptTokens),
				it.Equal(result.Usage.ReplyTokens, test.outputTokens),
			)
		})
	}
}

func TestDecoderContentTypes(t *testing.T) {
	t.Run("empty_content", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-empty",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: "",
					},
				},
			},
			Usage: usage{
				PromptTokens: 10,
				OutputTokens: 0,
				UsedTokens:   10,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(result).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 10,
					"replyTokens": 0
				},
				"content": [{}]
			}`),
		)
	})

	t.Run("json_response_content", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-json",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: `{"status": "success", "data": {"result": 42, "message": "Calculation complete"}}`,
					},
				},
			},
			Usage: usage{
				PromptTokens: 20,
				OutputTokens: 15,
				UsedTokens:   35,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(result).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 20,
					"replyTokens": 15
				},
				"content": [
					{
						"text": "regex:\\{\"status\"\\s*:\\s*\"success\".*\"data\"\\s*:\\s*\\{.*\"result\"\\s*:\\s*42.*\"message\"\\s*:\\s*\"Calculation complete\".*\\}.*\\}"
					}
				]
			}`),
		)
	})

	t.Run("unicode_special_characters", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-unicode",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: "Here are some special characters: é, ñ, 中文, 🚀, ∑, ∞, and emojis: 🎉🔥💡",
					},
				},
			},
			Usage: usage{
				PromptTokens: 15,
				OutputTokens: 20,
				UsedTokens:   35,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(result).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 15,
					"replyTokens": 20
				},
				"content": [
					{
						"text": "regex:Here are some special characters.*é.*ñ.*中文.*🚀.*∑.*∞.*emojis.*🎉🔥💡"
					}
				]
			}`),
		)
	})
}

func TestDecoderEdgeCasesAndErrorScenarios(t *testing.T) {
	t.Run("whitespace_only_content", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-whitespace",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: "   \n\t  \r\n  ",
					},
				},
			},
			Usage: usage{
				PromptTokens: 5,
				OutputTokens: 1,
				UsedTokens:   6,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(result).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 5,
					"replyTokens": 1
				},
				"content": [
					{
						"text": "regex:\\s+"
					}
				]
			}`),
		)
	})

	t.Run("all_zero_tokens", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-zero",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: "Response with zero token counts",
					},
				},
			},
			Usage: usage{
				PromptTokens: 0,
				OutputTokens: 0,
				UsedTokens:   0,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(result).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 0,
					"replyTokens": 0
				},
				"content": [
					{
						"text": "Response with zero token counts"
					}
				]
			}`),
		)
	})
}

func TestDecoderModelSpecificScenarios(t *testing.T) {
	t.Run("gpt4_turbo_response", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-abc123def456",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: `Based on my analysis of the code.`,
					},
				},
			},
			Usage: usage{
				PromptTokens: 180,
				OutputTokens: 95,
				UsedTokens:   275,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(result).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 180,
					"replyTokens": 95
				},
				"content": [
					{
						"text": "Based on my analysis of the code."
					}
				]
			}`),
		)
	})

	t.Run("gpt35_turbo_response", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-gpt35-response",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: "Sure! I'd be happy to help you with that. Could you please provide more details about what specific aspect you'd like me to focus on?",
					},
				},
			},
			Usage: usage{
				PromptTokens: 45,
				OutputTokens: 28,
				UsedTokens:   73,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			it.Json(result).Equiv(`{
				"stage": "return",
				"usage": {
					"inputTokens": 45,
					"replyTokens": 28
				},
				"content": [
					{
						"text": "regex:Sure.*I'd be happy to help.*Could you please provide more details.*specific aspect.*focus on\\?"
					}
				]
			}`),
		)
	})
}

func TestDecoderProtocolCompatibility(t *testing.T) {
	t.Run("openai_api_format_validation", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-standard-format",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: "This validates the standard OpenAI API response format.",
					},
				},
			},
			Usage: usage{
				PromptTokens: 12,
				OutputTokens: 8,
				UsedTokens:   20,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			// Validate that the result matches chatter.Reply structure
			it.True(result.Content != nil),
			it.Equal(len(result.Content), 1),
			it.True(result.Usage.InputTokens >= 0),
			it.True(result.Usage.ReplyTokens >= 0),
		)

		// Verify the content is properly typed as chatter.Text
		content, ok := result.Content[0].(chatter.Text)
		it.Then(t).Should(
			it.True(ok),
			it.Equal(string(content), "This validates the standard OpenAI API response format."),
		)
	})

	t.Run("field_mapping_accuracy", func(t *testing.T) {
		input := &reply{
			ID: "chatcmpl-mapping-test",
			Choices: []choice{
				{
					Message: message{
						Role:    "assistant",
						Content: "Field mapping verification",
					},
				},
			},
			Usage: usage{
				PromptTokens: 100,
				OutputTokens: 50,
				UsedTokens:   150,
			},
		}

		result, err := decoder{}.Decode(input)

		it.Then(t).Should(
			it.Nil(err),
			// Verify exact field mapping from OpenAI to chatter format
			it.Equal(result.Usage.InputTokens, input.Usage.PromptTokens),
			it.Equal(result.Usage.ReplyTokens, input.Usage.OutputTokens),
		)

		// Verify content extraction from first choice
		expectedContent := input.Choices[0].Message.Content
		actualContent := string(result.Content[0].(chatter.Text))
		it.Then(t).Should(
			it.Equal(actualContent, expectedContent),
		)
	})
}
