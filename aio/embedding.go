//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package aio

import (
	"context"
	"fmt"

	"github.com/kshard/chatter"
)

// Embedding is a wrapper for LLMs that support embeddings.
// It provides simple interface to get embeddings vectors for text.
type Embedder struct {
	chatter.Chatter
}

func NewEmbedder(chatter chatter.Chatter) *Embedder {
	return &Embedder{
		Chatter: chatter,
	}
}

func (api *Embedder) Embedding(ctx context.Context, text string) ([]float32, int, error) {
	reply, err := api.Chatter.Prompt(ctx,
		[]chatter.Message{chatter.Text(text)},
	)
	if err != nil {
		return nil, 0, err
	}

	for _, content := range reply.Content {
		switch c := content.(type) {
		case chatter.Vector:
			return c, reply.Usage.InputTokens + reply.Usage.ReplyTokens, nil
		}
	}

	return nil, 0, fmt.Errorf("invalid response, no vector found")
}
