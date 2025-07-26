//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llama

import (
	"testing"

	"github.com/fogfish/it/v2"
)

func TestDecoderSuccessfulCompletion(t *testing.T) {
	input := &reply{
		Text:             "The capital of France is Paris.",
		UsedPromptTokens: 15,
		UsedTextTokens:   8,
		StopReason:       "stop",
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
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

func TestDecoderIncompleteResponse(t *testing.T) {
	input := &reply{
		Text:             "The capital of France is",
		UsedPromptTokens: 10,
		UsedTextTokens:   5,
		StopReason:       "length",
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
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
}

func TestDecoderErrorResponse(t *testing.T) {
	input := &reply{
		Text:             "",
		UsedPromptTokens: 5,
		UsedTextTokens:   0,
		StopReason:       "error",
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
			"stage": "error",
			"usage": {
				"inputTokens": 5,
				"replyTokens": 0
			},
			"content": [{}]
		}`),
	)
}

func TestDecoderUnknownStopReason(t *testing.T) {
	input := &reply{
		Text:             "Partial response",
		UsedPromptTokens: 12,
		UsedTextTokens:   2,
		StopReason:       "unknown_reason",
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
			"stage": "error",
			"usage": {
				"inputTokens": 12,
				"replyTokens": 2
			},
			"content": [
				{"text": "Partial response"}
			]
		}`),
	)
}

func TestDecoderTokenUsageAccuracy(t *testing.T) {
	input := &reply{
		Text:             "This is a longer response that would use more tokens for testing purposes.",
		UsedPromptTokens: 25,
		UsedTextTokens:   18,
		StopReason:       "stop",
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 25,
				"replyTokens": 18
			},
			"content": [
				{"text": "This is a longer response that would use more tokens for testing purposes."}]
		}`),
	)
}

func TestDecoderEmptyTextWithStopReason(t *testing.T) {
	input := &reply{
		Text:             "",
		UsedPromptTokens: 8,
		UsedTextTokens:   0,
		StopReason:       "stop",
	}

	result, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(result).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 8,
				"replyTokens": 0
			},
			"content": [{}]
		}`),
	)
}
