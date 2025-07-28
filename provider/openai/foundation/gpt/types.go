//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package gpt

import (
	"github.com/fogfish/logger/x/xlog"
	"github.com/kshard/chatter/aio/provider"
	"github.com/kshard/chatter/provider/openai"
)

// See https://platform.openai.com/docs/api-reference/chat/create

type input struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type reply struct {
	ID      string   `json:"id"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Message message `json:"message"`
}

type usage struct {
	PromptTokens int `json:"prompt_tokens"`
	OutputTokens int `json:"completion_tokens"`
	UsedTokens   int `json:"total_tokens"`
}

type encoder struct{ req input }

type decoder struct{}

type GPT = provider.Provider[*input, *reply]

func New(model string, opt ...openai.Option) (*GPT, error) {
	service, err := openai.New[*input, *reply]("/v1/chat/completions", opt...)
	if err != nil {
		return nil, err
	}

	return provider.New(factory(model), decoder{}, service), nil
}

func Must[T any](api T, err error) T {
	if err != nil {
		xlog.Emergency("openai gpt model has failed", err)
	}
	return api
}
