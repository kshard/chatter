//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package aio

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/gob"
	"fmt"
	"log/slog"

	"github.com/kshard/chatter"
)

// Getter interface abstract storage
type Getter interface{ Get([]byte) ([]byte, error) }

// Setter interface abstract storage
type Putter interface{ Put([]byte, []byte) error }

type Eraser interface{ Delete([]byte) error }

// KeyVal interface
type KeyVal interface {
	Getter
	Putter
	Eraser
}

// Caching strategy for LLMs I/O
type Cache struct {
	chatter.Chatter
	cache KeyVal
}

var _ chatter.Chatter = (*Cache)(nil)

func init() {
	gob.Register(chatter.Text(""))
	gob.Register(chatter.Vector{})
	gob.Register([]chatter.Content{})
}

// Creates read-through caching layer for LLM client.
//
// Use github.com/akrylysov/pogreb to cache chatter on local file systems:
//
//	llm, err := /* create LLM client */
//	db, err := pogreb.Open("llm.cache", nil)
//	text := aio.NewCache(db, llm)
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

func (c *Cache) Prompt(ctx context.Context, prompt []chatter.Message, opts ...chatter.Opt) (*chatter.Reply, error) {
	if len(prompt) == 0 {
		return nil, fmt.Errorf("bad request, empty prompt")
	}

	hkey := c.HashKey(prompt[len(prompt)-1].String())
	val, err := c.cache.Get(hkey)
	if err != nil {
		return nil, err
	}

	if len(val) != 0 {
		reply, err := decode(val)
		if err == nil {
			return reply, nil
		}

		slog.Warn("failed to decode cached LLM reply", "err", err)
		c.cache.Delete(hkey)
	}

	reply, err := c.Chatter.Prompt(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}

	if reply.Stage == chatter.LLM_RETURN {
		bin, err := encode(reply)
		switch {
		case err != nil:
			slog.Warn("failed to encode LLM reply", "err", err)
			return reply, nil
		case bin == nil:
			return reply, nil
		default:
			err = c.cache.Put(hkey, bin)
			if err != nil {
				slog.Warn("failed to cache LLM reply", "err", err)
			}
			return reply, nil
		}
	}

	return reply, nil
}

func encode(reply *chatter.Reply) ([]byte, error) {
	if reply == nil {
		return nil, nil
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(&reply.Content); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func decode(data []byte) (*chatter.Reply, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(data))

	var val []chatter.Content
	if err := dec.Decode(&val); err != nil {
		return nil, err
	}

	return &chatter.Reply{
		Stage:   chatter.LLM_RETURN,
		Content: val,
	}, nil
}
