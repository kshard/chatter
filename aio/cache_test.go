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
	c := aio.NewCache(kv, mock{
		&chatter.Reply{
			Stage: chatter.LLM_RETURN,
			Content: []chatter.Content{
				chatter.Text("test"),
				chatter.Vector{1.0, 2.0, 3.0},
			},
		},
	})

	var prompt chatter.Prompt
	prompt.WithTask("Make me a test.")

	c.Prompt(context.Background(), prompt.ToSeq())
	reply, err := c.Prompt(context.Background(), prompt.ToSeq())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.Stage != chatter.LLM_RETURN {
		t.Fatalf("unexpected stage: %s", reply.Stage)
	}
	if len(reply.Content) != 2 {
		t.Fatalf("unexpected content length: %d", len(reply.Content))
	}
	if reply.Content[0].(chatter.Text) != "test" {
		t.Fatalf("unexpected content[0]: %v", reply.Content[0])
	}
	if v, ok := reply.Content[1].(chatter.Vector); !ok || len(v) != 3 || v[0] != 1.0 || v[1] != 2.0 || v[2] != 3.0 {
		t.Fatalf("unexpected content[1]: %v", reply.Content[1])
	}

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
