package titan

import (
	"testing"

	"github.com/fogfish/it/v2"
)

func TestTitanDecoderBasicEmbedding(t *testing.T) {
	input := &reply{
		Vector:         []float32{0.1, 0.2, 0.3, 0.4, 0.5},
		UsedTextTokens: 25,
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

func TestTitanDecoderEmptyVector(t *testing.T) {
	input := &reply{
		Vector:         []float32{},
		UsedTextTokens: 0,
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

func TestTitanDecoderZeroTokenUsage(t *testing.T) {
	input := &reply{
		Vector:         []float32{-0.5, 0.0, 0.5},
		UsedTextTokens: 0,
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

func TestTitanDecoderNegativeValues(t *testing.T) {
	input := &reply{
		Vector:         []float32{-1.0, -0.5, 0.0, 0.5, 1.0},
		UsedTextTokens: 42,
	}

	reply, err := decoder{}.Decode(input)

	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {
				"inputTokens": 42,
				"replyTokens": 0
			},
			"content": [
				{
					"vector": [-1.0, -0.5, 0.0, 0.5, 1.0]
				}
			]
		}`),
	)
}

func TestTitanDecoderHighPrecisionValues(t *testing.T) {
	input := &reply{
		Vector:         []float32{0.123456789, 0.987654321, -0.111111111},
		UsedTextTokens: 100,
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
					"vector": [0.12345679,0.9876543,-0.11111111]
				}
			]
		}`),
	)
}
