//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import (
	"encoding/json"
	"testing"

	"github.com/fogfish/it/v2"
)

func TestWireDecoderBasicTextResponse(t *testing.T) {
	// https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_Converse.html#API_runtime_Converse_Example_2
	var out wireOutput
	err := json.Unmarshal([]byte(`{
		"output": {
			"message": {
				"role": "assistant",
				"content": [{"text": "generated text"}]
			}
		},
		"stopReason": "end_turn",
		"usage": {"inputTokens": 30, "outputTokens": 628, "totalTokens": 658}
	}`), &out)
	it.Then(t).Must(it.Nil(err))

	reply, err := wireDecoder{}.Decode(&out)
	it.Then(t).Should(
		it.Nil(err),
		it.Json(reply).Equiv(`{
			"stage": "return",
			"usage": {"inputTokens": 30, "replyTokens": 628},
			"content": [{"text": "generated text"}]
		}`),
	)
}

func TestWireDecoderToolUseResponse(t *testing.T) {
	var out wireOutput
	err := json.Unmarshal([]byte(`{
		"output": {
			"message": {
				"role": "assistant",
				"content": [{"toolUse": {"toolUseId": "tool-1", "name": "get_weather", "input": {"city": "Berlin"}}}]
			}
		},
		"stopReason": "tool_use",
		"usage": {"inputTokens": 10, "outputTokens": 5, "totalTokens": 15}
	}`), &out)
	it.Then(t).Must(it.Nil(err))

	reply, err := wireDecoder{}.Decode(&out)
	it.Then(t).Should(
		it.Nil(err),
		it.Equal(reply.Stage, "invoke"),
		it.Equal(len(reply.Content), 1),
	)
}

func TestWireDecoderStageMapping(t *testing.T) {
	for reason, stage := range map[string]string{
		"end_turn":         "return",
		"max_tokens":       "incomplete",
		"stop_sequence":    "incomplete",
		"tool_use":         "invoke",
		"content_filtered": "error",
	} {
		reply, err := wireDecoder{}.Decode(&wireOutput{StopReason: reason})
		it.Then(t).Should(
			it.Nil(err),
			it.Equal(string(reply.Stage), stage),
		)
	}
}
