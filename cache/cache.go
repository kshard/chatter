//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package cache

import (
	"context"
	"crypto/sha1"
	"encoding"
	"log/slog"

	"github.com/kshard/chatter"
)

// Creates caching layer for LLM client.
//
// Use github.com/akrylysov/pogreb to cache chatter on local file systems:
//
//	llm, err := /* create LLM client */
//	db, err := pogreb.Open("llm.cache", nil)
//	text := cache.New(db, llm)
func New(cache Cache, chatter chatter.Chatter) *Client {
	return &Client{
		Chatter: chatter,
		cache:   cache,
	}
}

func (c *Client) HashKey(prompt encoding.TextMarshaler) ([]byte, error) {
	b, err := prompt.MarshalText()
	if err != nil {
		return nil, err
	}

	hash := sha1.New()
	hash.Write(b)
	return hash.Sum(nil), nil
}

func (c *Client) Prompt(ctx context.Context, prompt encoding.TextMarshaler, opts ...func(*chatter.Options)) (string, error) {
	hkey, err := c.HashKey(prompt)
	if err != nil {
		return "", err
	}

	val, err := c.cache.Get(hkey)
	if err != nil {
		return "", err
	}

	if len(val) != 0 {
		return string(val), nil
	}

	reply, err := c.Chatter.Prompt(ctx, prompt, opts...)
	if err != nil {
		return "", err
	}

	err = c.cache.Put(hkey, []byte(reply))
	if err != nil {
		slog.Warn("failed to cache LLM reply", "err", err)
	}

	return reply, nil
}
