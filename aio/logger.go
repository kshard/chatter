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
	"io"

	"github.com/kshard/chatter"
)

// Logger of LLM's I/O
type Logger struct {
	chatter.Chatter
	w io.Writer
}

// Create new debugger session
func NewLogger(w io.Writer, chatter chatter.Chatter) *Logger {
	return &Logger{
		Chatter: chatter,
		w:       w,
	}
}

func (deb *Logger) Prompt(ctx context.Context, seq []fmt.Stringer, opt ...chatter.Opt) (chatter.Reply, error) {
	if len(seq) != 0 {
		ask := seq[len(seq)-1].String()
		deb.w.Write([]byte("\n>>>>>\n" + ask + "\n"))
	}

	reply, err := deb.Chatter.Prompt(ctx, seq, opt...)
	if err != nil {
		deb.w.Write([]byte("FAIL:\n\t" + err.Error() + "\n"))
	} else {
		deb.w.Write([]byte("\n<<<<<\n" + reply.Text + "\n"))
	}

	return reply, err
}
