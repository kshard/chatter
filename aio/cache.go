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
	"crypto/sha1"
	"fmt"
	"log/slog"

	"github.com/kshard/chatter"
)

// Getter interface abstract storage
type Getter interface{ Get([]byte) ([]byte, error) }

// Setter interface abstract storage
type Putter interface{ Put([]byte, []byte) error }

// KeyVal interface
type KeyVal interface {
	Getter
	Putter
}

// Caching strategy for LLMs I/O
type Cache struct {
	chatter.Chatter
	cache KeyVal
}

var _ chatter.Chatter = (*Cache)(nil)

// Creates read-through caching layer for LLM client.
//
// Use github.com/akrylysov/pogreb to cache chatter on local file systems:
//
//	llm, err := /* create LLM client */
//	db, err := pogreb.Open("llm.cache", nil)
//	text := cache.New(db, llm)
func NewCache(cache KeyVal, chatter chatter.Chatter) *Cache {
	return &Cache{
		Chatter: chatter,
		cache:   cache,
	}
}

func (c *Cache) HashKey(prompt string) []byte {
	hash := sha1.New()
	hash.Write([]byte(prompt))
	return hash.Sum(nil)
}

func (c *Cache) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (chatter.Reply, error) {
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
