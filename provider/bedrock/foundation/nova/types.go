//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package nova

import (
	"github.com/kshard/chatter/aio/provider"
	"github.com/kshard/chatter/provider/bedrock"
	"github.com/kshard/chatter/provider/bedrock/batch"
)

// Nova I/O Schema:
// https://docs.aws.amazon.com/nova/latest/userguide/complete-request-schema.html

type input struct {
	System          []content        `json:"system,omitempty"`
	Messages        []message        `json:"messages"`
	InferenceConfig *inferenceConfig `json:"inferenceConfig,omitempty"`

	// Note: tools are not enabled use Converse API for this purpose
}

type message struct {
	Role    string    `json:"role"`
	Content []content `json:"content"`
}

type content struct {
	Text  string `json:"text,omitempty"`
	Image any    `json:"image,omitempty"`
	Video any    `json:"video,omitempty"`
}

type inferenceConfig struct {
	MaxTokens     int      `json:"maxTokens,omitempty"`
	Temperature   float64  `json:"temperature,omitempty"`
	TopP          float64  `json:"topP,omitempty"`
	TopK          int      `json:"topK,omitempty"`
	StopSequences []string `json:"stopSequences,omitempty"`
}

type reply struct {
	Output     output `json:"output"`
	Usage      usage  `json:"usage"`
	StopReason string `json:"stop_reason"`
}

type output struct {
	Message message `json:"message"`
}

type usage struct {
	InputTokens               int `json:"inputTokens,omitempty"`
	OutputTokens              int `json:"outputTokens,omitempty"`
	TotalTokens               int `json:"totalTokens,omitempty"`
	CacheReadInputTokenCount  int `json:"cacheReadInputTokenCount,omitempty"`
	CacheWriteInputTokenCount int `json:"cacheWriteInputTokenCount,omitempty"`
}

type encoder struct {
	req input
}

type decoder struct{}

type Nova = provider.Provider[*input, *reply]

func New(model string, opts ...bedrock.Option) (*Nova, error) {
	service, err := bedrock.New[*input, *reply](model, opts...)
	if err != nil {
		return nil, err
	}

	return provider.New(factory, decoder{}, service), nil
}

type NovaBatch = batch.Provider[*input, *reply]

func NewBatch(fs *batch.FileSystem, model string, opts ...batch.Option) (*NovaBatch, error) {
	return batch.New(fs, model, factory, decoder{}, opts...)
}
