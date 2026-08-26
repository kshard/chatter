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
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func TestWireEncoderMatchesConverseWireSchema(t *testing.T) {
	// https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_Converse.html#API_runtime_Converse_Example_1
	f, err := wireFactory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsStratum(chatter.Stratum("You are an economist with access to lots of data"))
	it.Then(t).Must(it.Nil(err))

	err = f.AsText(chatter.Text("Write an article about impact of high inflation to GDP of a country"))
	it.Then(t).Must(it.Nil(err))

	f.WithInferrer(provider.Inferrer{MaxTokens: 1000, Temperature: 0.5})

	it.Then(t).Should(
		it.Json(f.Build()).Equiv(`{
			"system": [{"text": "You are an economist with access to lots of data"}],
			"messages": [
				{
					"role": "user",
					"content": [{"text": "Write an article about impact of high inflation to GDP of a country"}]
				}
			],
			"inferenceConfig": {"maxTokens": 1000, "temperature": 0.5}
		}`),
	)
}

func TestWireEncoderToolConfig(t *testing.T) {
	f, err := wireFactory()
	it.Then(t).Must(it.Nil(err))

	f.WithCommand(chatter.Cmd{
		Cmd:    "get_weather",
		About:  "Returns the current weather for a city",
		Schema: json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
	})

	req := f.Build()
	it.Then(t).Should(
		it.Nil(err),
		it.Equal(len(req.ToolConfig.Tools), 1),
		it.Equal(req.ToolConfig.Tools[0].ToolSpec.Name, "get_weather"),
		it.True(req.ToolConfig.ToolChoice.Auto != nil),
	)
}

func TestWireEncoderAnswerProducesToolResult(t *testing.T) {
	f, err := wireFactory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsAnswer(&chatter.Answer{
		Yield: []chatter.Json{{ID: "tool-1", Value: json.RawMessage(`{"temp": 72}`)}},
	})
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(
		it.Json(f.Build()).Equiv(`{
			"messages": [
				{
					"role": "user",
					"content": [
						{"toolResult": {"toolUseId": "tool-1", "content": [{"json": {"temp": 72}}]}}
					]
				}
			]
		}`),
	)
}

func TestWireEncoderReplyProducesAssistantMessage(t *testing.T) {
	f, err := wireFactory()
	it.Then(t).Must(it.Nil(err))

	err = f.AsReply(&chatter.Reply{
		Content: []chatter.Content{
			chatter.Text("let me check"),
			chatter.Invoke{Cmd: "get_weather", Args: chatter.Json{ID: "tool-1", Value: json.RawMessage(`{"city":"Berlin"}`)}},
		},
	})
	it.Then(t).Must(it.Nil(err))

	it.Then(t).Should(
		it.Json(f.Build()).Equiv(`{
			"messages": [
				{
					"role": "assistant",
					"content": [
						{"text": "let me check"},
						{"toolUse": {"toolUseId": "tool-1", "name": "get_weather", "input": {"city": "Berlin"}}}
					]
				}
			]
		}`),
	)
}
