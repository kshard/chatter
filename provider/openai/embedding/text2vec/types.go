//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text2vec

import (
	"strings"

	"github.com/kshard/chatter/aio/provider"
	"github.com/kshard/chatter/provider/openai"
)

type input struct {
	Model      string `json:"model"`
	Text       string `json:"input"`
	Dimensions int    `json:"dimensions,omitempty"`
}

type reply struct {
	Object  string   `json:"object"`
	Vectors []vector `json:"data"`
	Model   string   `json:"model"`
	Usage   usage    `json:"usage"`
}

type vector struct {
	Object string    `json:"object"`
	Index  int       `json:"index"`
	Vector []float32 `json:"embedding"`
}

type usage struct {
	PromptTokens int `json:"prompt_tokens"`
	UsedTokens   int `json:"total_tokens"`
}

type encoder struct {
	w   strings.Builder
	req input
}

type decoder struct{}

type Text = provider.Provider[*input, *reply]

func New(model string, dimensions int, opts ...openai.Option) (*Text, error) {
	service, err := openai.New[*input, *reply]("/v1/embeddings", opts...)
	if err != nil {
		return nil, err
	}

	return provider.New(factory(model, dimensions), decoder{}, service), nil
}
