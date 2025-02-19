//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/embeddings
//

package cache_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/cache"
)

func TestCache(t *testing.T) {
	kv := keyval{}
	c := cache.New(kv, llm{})

	var prompt chatter.Prompt
	prompt.WithTask("Make me a test.")

	c.Prompt(context.Background(), prompt.ToSeq())
	c.Prompt(context.Background(), prompt.ToSeq())

	for k := range kv {
		hkey := c.HashKey(prompt.String())
		if !bytes.Equal([]byte(k), hkey) {
			t.Errorf("unexpected key")
		}
	}
}

// mock embedding client
type llm struct{}

func (llm) UsedInputTokens() int { return 5 }
func (llm) UsedReplyTokens() int { return 10 }

func (llm) Prompt(context.Context, []fmt.Stringer, ...chatter.Opt) (chatter.Reply, error) {
	return chatter.Reply{Text: "Looking for testing"}, nil
}

// mock key-value
type keyval map[string][]byte

func (kv keyval) Get(key []byte) ([]byte, error) {
	if val, has := kv[string(key)]; has {
		return val, nil
	}

	return nil, nil
}

// Setter interface abstract storage
func (kv keyval) Put(key []byte, val []byte) error {
	kv[string(key)] = val
	return nil
}
