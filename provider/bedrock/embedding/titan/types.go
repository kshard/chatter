//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package titan

import (
	"strings"

	"github.com/kshard/chatter/aio/provider"
	"github.com/kshard/chatter/provider/bedrock"
)

type input struct {
	Text       string `json:"inputText"`
	Dimensions int    `json:"dimensions,omitempty"`
}

type reply struct {
	Vector         []float32 `json:"embedding"`
	UsedTextTokens int       `json:"inputTextTokenCount"`
}

type encoder struct {
	w   strings.Builder
	req input
}

type decoder struct{}

const (
	EMBEDDING_SIZE_256  = 256
	EMBEDDING_SIZE_512  = 512
	EMBEDDING_SIZE_1024 = 1024
)

type Titan = provider.Provider[*input, *reply]

func New(model string, size int, opts ...bedrock.Option) (*Titan, error) {
	service, err := bedrock.New[*input, *reply](model, opts...)
	if err != nil {
		return nil, err
	}

	return provider.New(factory(size), decoder{}, service), nil
}
