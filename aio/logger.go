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
	"encoding/json"
	"io"

	"github.com/kshard/chatter"
)

// Logger of LLM's I/O
type Logger struct {
	chatter.Chatter
	w          io.Writer
	jsonFormat bool
}

func NewTextLogger(w io.Writer, chatter chatter.Chatter) *Logger {
	return &Logger{
		Chatter:    chatter,
		w:          w,
		jsonFormat: false,
	}
}

func NewJsonLogger(w io.Writer, chatter chatter.Chatter) *Logger {
	return &Logger{
		Chatter:    chatter,
		w:          w,
		jsonFormat: true,
	}
}

func (deb *Logger) Prompt(ctx context.Context, seq []chatter.Message, opt ...chatter.Opt) (*chatter.Reply, error) {
	if len(seq) != 0 {
		ask := seq[len(seq)-1]
		deb.logEgress(ask)
	}

	reply, err := deb.Chatter.Prompt(ctx, seq, opt...)
	if err != nil {
		deb.w.Write([]byte("FAIL:\n\t" + err.Error() + "\n"))
	} else {
		deb.logIngress(reply)
	}

	return reply, err
}

func (deb *Logger) logEgress(msg chatter.Message) {
	deb.w.Write([]byte("\n>>>>>\n"))

	if !deb.jsonFormat {
		deb.w.Write([]byte(msg.String()))
		deb.w.Write([]byte("\n"))
		return
	}

	b, err := json.MarshalIndent(msg, "|", "  ")
	if err == nil {
		deb.w.Write([]byte("|"))
		deb.w.Write(b)
		deb.w.Write([]byte("\n"))
	}
}

func (deb *Logger) logIngress(msg *chatter.Reply) {
	deb.w.Write([]byte("\n<<<<<\n"))

	if !deb.jsonFormat {
		deb.w.Write([]byte(msg.String()))
		deb.w.Write([]byte("\n"))
		return
	}

	b, err := json.MarshalIndent(msg, "|", "  ")
	if err == nil {
		deb.w.Write([]byte("|"))
		deb.w.Write(b)
		deb.w.Write([]byte("\n"))
	}
}
