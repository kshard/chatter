//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/embeddings
//

package aio_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio"
)

func TestCache(t *testing.T) {
	kv := keyval{}
	c := aio.NewCache(kv, mock{})

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
