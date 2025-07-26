//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text2vec

import (
	"testing"

	"github.com/fogfish/it/v2"
)

func TestDecoderBasicEmbedding(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-small",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{0.1, 0.2, 0.3, 0.4, 0.5},
			},
		},
		Usage: usage{
			PromptTokens: 10,
			UsedTokens:   25,
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 25,
				"replyTokens": 0
			},
			"content": [
				{
					"vector": [0.1, 0.2, 0.3, 0.4, 0.5]
				}
			]
		}`),
	)
}

func TestDecoderLargeEmbeddingVector(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-large",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{-0.5, 0.0, 0.5, 1.0, -1.0, 0.25, -0.25, 0.75, -0.75, 0.125},
			},
		},
		Usage: usage{
			PromptTokens: 50,
			UsedTokens:   100,
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 100,
				"replyTokens": 0
			},
			"content": [
				{
					"vector": [-0.5, 0.0, 0.5, 1.0, -1.0, 0.25, -0.25, 0.75, -0.75, 0.125]
				}
			]
		}`),
	)
}

func TestDecoderHighPrecisionValues(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-ada-002",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{0.123456789, 0.987654321, -0.111111111, 0.555555555},
			},
		},
		Usage: usage{
			PromptTokens: 75,
			UsedTokens:   200,
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 200,
				"replyTokens": 0
			},
			"content": [
				{
					"vector": [0.12345679, 0.9876543, -0.11111111, 0.5555556]
				}
			]
		}`),
	)
}

func TestDecoderEmptyVector(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-small",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{},
			},
		},
		Usage: usage{
			PromptTokens: 0,
			UsedTokens:   0,
		},
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
			"content": [{}]
		}`),
	)
}

func TestDecoderZeroTokenUsage(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-small",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{-0.5, 0.0, 0.5},
			},
		},
		Usage: usage{
			PromptTokens: 5,
			UsedTokens:   0,
		},
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
					"vector": [-0.5, 0.0, 0.5]
				}
			]
		}`),
	)
}

func TestDecoderInvalidResponseNoVectors(t *testing.T) {
	input := &reply{
		Object:  "list",
		Model:   "text-embedding-3-small",
		Vectors: []vector{},
		Usage: usage{
			PromptTokens: 10,
			UsedTokens:   25,
		},
	}

	it.Then(t).Should(
		it.Error(decoder{}.Decode(input)).Contain("invalid response"),
	)
}

func TestDecoderInvalidResponseMultipleVectors(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-small",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{0.1, 0.2, 0.3},
			},
			{
				Object: "embedding",
				Index:  1,
				Vector: []float32{0.4, 0.5, 0.6},
			},
		},
		Usage: usage{
			PromptTokens: 20,
			UsedTokens:   50,
		},
	}

	it.Then(t).Should(
		it.Error(decoder{}.Decode(input)).Contain("invalid response"),
	)
}

func TestDecoderEdgeCaseExtremeValues(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-large",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{-999.999999, 0.0, 999.999999, -0.000001, 0.000001},
			},
		},
		Usage: usage{
			PromptTokens: 1000,
			UsedTokens:   1500,
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 1500,
				"replyTokens": 0
			},
			"content": [
				{
					"vector": [-1000.0, 0.0, 1000.0, -0.000001, 0.000001]
				}
			]
		}`),
	)
}

func TestDecoderVectorIndexValidation(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-small",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  99,
				Vector: []float32{0.1, 0.2, 0.3},
			},
		},
		Usage: usage{
			PromptTokens: 15,
			UsedTokens:   30,
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 30,
				"replyTokens": 0
			},
			"content": [
				{
					"vector": [0.1, 0.2, 0.3]
				}
			]
		}`),
	)
}

func TestDecoderCompleteMetadataResponse(t *testing.T) {
	input := &reply{
		Object: "list",
		Model:  "text-embedding-3-small",
		Vectors: []vector{
			{
				Object: "embedding",
				Index:  0,
				Vector: []float32{0.707, -0.707, 0.0, 1.0, -1.0},
			},
		},
		Usage: usage{
			PromptTokens: 42,
			UsedTokens:   84,
		},
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 84,
				"replyTokens": 0
			},
			"content": [
				{
					"vector": [0.707, -0.707, 0.0, 1.0, -1.0]
				}
			]
		}`),
	)
}
