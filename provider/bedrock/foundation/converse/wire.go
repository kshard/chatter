//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import "encoding/json"

// Plain-JSON mirror of the Converse API's wire schema, used to build
// records for Bedrock Batch Inference (modelInvocationType: Converse).
//
// The types above (encoder/decoder in encoder.go, decoder.go) build
// *bedrockruntime.ConverseInput/Output, the AWS SDK's Go representation
// of Converse. Those types are Smithy unions with no JSON struct tags:
// the SDK serializes them through its own protocol codec when making a
// live Converse call, not through encoding/json. Batch inference has no
// such codec in the loop — each record is JSON bytes written directly to
// a JSONL file — so a request built from *bedrockruntime.ConverseInput
// cannot be reused there. The types below encode/decode the literal
// Converse REST JSON schema instead, independent of the SDK's Go types,
// so they work equally for any model that Bedrock Batch accepts through
// the Converse format.
// https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_Converse.html
type wireInput struct {
	Messages        []wireMessage        `json:"messages"`
	System          []wireContent        `json:"system,omitempty"`
	InferenceConfig *wireInferenceConfig `json:"inferenceConfig,omitempty"`
	ToolConfig      *wireToolConfig      `json:"toolConfig,omitempty"`
}

type wireMessage struct {
	Role    string        `json:"role"`
	Content []wireContent `json:"content"`
}

// wireContent is a Converse content block. Bedrock discriminates the
// block's kind by which field is present, not by an explicit type tag.
type wireContent struct {
	Text       string          `json:"text,omitempty"`
	ToolUse    *wireToolUse    `json:"toolUse,omitempty"`
	ToolResult *wireToolResult `json:"toolResult,omitempty"`
}

type wireToolUse struct {
	ToolUseID string          `json:"toolUseId"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
}

type wireToolResult struct {
	ToolUseID string                  `json:"toolUseId"`
	Content   []wireToolResultContent `json:"content"`
}

type wireToolResultContent struct {
	JSON json.RawMessage `json:"json,omitempty"`
	Text string          `json:"text,omitempty"`
}

type wireInferenceConfig struct {
	MaxTokens     int      `json:"maxTokens,omitempty"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"topP,omitempty"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

type wireToolConfig struct {
	ToolChoice wireToolChoice `json:"toolChoice"`
	Tools      []wireTool     `json:"tools"`
}

type wireToolChoice struct {
	Auto *struct{} `json:"auto,omitempty"`
}

type wireTool struct {
	ToolSpec wireToolSpec `json:"toolSpec"`
}

type wireToolSpec struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	InputSchema wireToolInputSchema `json:"inputSchema"`
}

type wireToolInputSchema struct {
	JSON json.RawMessage `json:"json"`
}

type wireOutput struct {
	Output     *wireOutputMessage `json:"output,omitempty"`
	StopReason string             `json:"stopReason"`
	Usage      wireUsage          `json:"usage"`
}

type wireOutputMessage struct {
	Message wireMessage `json:"message"`
}

type wireUsage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}
