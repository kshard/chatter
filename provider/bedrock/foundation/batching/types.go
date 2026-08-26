//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

// Package batching implements AWS Bedrock Batch Inference using the
// Converse API's wire format (modelInvocationType: Converse), so it works
// with any model Bedrock accepts through Converse — Claude, Nova, Llama,
// etc. — by changing only the model ID passed to [New].
//
// This is a separate package from foundation/converse on purpose: that
// package builds *bedrockruntime.ConverseInput/Output, the AWS SDK's Go
// representation of Converse used for live (real-time) calls, which the
// SDK serializes through its own protocol codec. Batch inference has no
// such codec in the loop — each record is JSON bytes written directly to
// a JSONL file — and *bedrockruntime.ConverseInput doesn't marshal to the
// correct wire shape via encoding/json (its union types carry no JSON
// tags; marshaling one produces field names like "Value" and "ModelId"
// instead of "text" and no top-level modelId at all). The types below
// encode/decode the literal Converse REST JSON schema instead, independent
// of the SDK's Go types, so live and batch concerns don't mix.
// https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_Converse.html
package batching

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/kshard/chatter/provider/bedrock/batch"
)

type input struct {
	Messages        []message        `json:"messages"`
	System          []content        `json:"system,omitempty"`
	InferenceConfig *inferenceConfig `json:"inferenceConfig,omitempty"`
	ToolConfig      *toolConfig      `json:"toolConfig,omitempty"`
}

type message struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

// content is a Converse content block. Bedrock discriminates the block's
// kind by which field is present, not by an explicit type tag.
type content struct {
	Text       string      `json:"text,omitempty"`
	ToolUse    *toolUse    `json:"toolUse,omitempty"`
	ToolResult *toolResult `json:"toolResult,omitempty"`
}

type toolUse struct {
	ToolUseID string          `json:"toolUseId"`
	Name      string          `json:"name"`
	Input     json.RawMessage `json:"input"`
}

type toolResult struct {
	ToolUseID string              `json:"toolUseId"`
	Content   []toolResultContent `json:"content"`
}

type toolResultContent struct {
	JSON json.RawMessage `json:"json,omitempty"`
	Text string          `json:"text,omitempty"`
}

type inferenceConfig struct {
	MaxTokens     int      `json:"maxTokens,omitempty"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"topP,omitempty"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

type toolConfig struct {
	ToolChoice toolChoice `json:"toolChoice"`
	Tools      []tool     `json:"tools"`
}

type toolChoice struct {
	Auto *struct{} `json:"auto,omitempty"`
}

type tool struct {
	ToolSpec toolSpec `json:"toolSpec"`
}

type toolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema toolInputSchema `json:"inputSchema"`
}

type toolInputSchema struct {
	JSON json.RawMessage `json:"json"`
}

type output struct {
	Output     *outputMessage `json:"output,omitempty"`
	StopReason string         `json:"stopReason"`
	Usage      usage          `json:"usage"`
}

type outputMessage struct {
	Message message `json:"message"`
}

type usage struct {
	InputTokens  int `json:"inputTokens"`
	OutputTokens int `json:"outputTokens"`
}

// Batch is a Bedrock Batch Inference job that encodes records in the
// Converse API's wire format.
type Batch = batch.Provider[*input, *output]

// New prepares a Batch Inference job for model, using the Converse API's
// wire format (modelInvocationType: Converse). Any model Bedrock accepts
// through Converse works here — Claude, Nova, Llama, etc. — by changing
// only the model ID.
func New(fs *batch.FileSystem, model string, opts ...batch.Option) (*Batch, error) {
	opts = append(opts, batch.WithModelInvocationType(types.ModelInvocationTypeConverse))
	return batch.New(fs, model, factory, decoder{}, opts...)
}
