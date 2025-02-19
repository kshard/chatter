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
	"fmt"
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

func (c *Client) HashKey(prompt string) []byte {
	hash := sha1.New()
	hash.Write([]byte(prompt))
	return hash.Sum(nil)
}

func (c *Client) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (chatter.Reply, error) {
	if len(prompt) == 0 {
		return chatter.Reply{}, fmt.Errorf("bad request, empty prompt")
	}

	hkey := c.HashKey(prompt[len(prompt)-1].String())
	val, err := c.cache.Get(hkey)
	if err != nil {
		return chatter.Reply{}, err
	}

	if len(val) != 0 {
		return chatter.Reply{Text: string(val)}, nil
	}

	reply, err := c.Chatter.Prompt(ctx, prompt, opts...)
	if err != nil {
		return chatter.Reply{}, err
	}

	err = c.cache.Put(hkey, []byte(reply.Text))
	if err != nil {
		slog.Warn("failed to cache LLM reply", "err", err)
	}

	return reply, nil
}
