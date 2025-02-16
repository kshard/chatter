//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llama3

import (
	"io"
)

const (
	begin_of_text   = "<|begin_of_text|>"
	start_header_id = "\n<|start_header_id|>"
	end_header_id   = "<|end_header_id|>\n"
	end_of_turn     = "\n<|eot_id|>\n"
	system          = "system"
	assistant       = "assistant"
	human           = "user"
)

type Encoder struct {
	w    io.Writer
	role string
	seq  int
}

func NewEncoder(w io.Writer, role string) (*Encoder, error) {
	codec := &Encoder{w: w, role: role, seq: 0}
	if err := codec.session(); err != nil {
		return nil, err
	}

	return codec, nil
}

func (e *Encoder) session() error {
	if _, err := e.w.Write([]byte(begin_of_text)); err != nil {
		return err
	}

	if err := e.stratum(); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) header(actor string) error {
	if _, err := e.w.Write([]byte(start_header_id)); err != nil {
		return err
	}
	if _, err := e.w.Write([]byte(actor)); err != nil {
		return err
	}
	if _, err := e.w.Write([]byte(end_header_id)); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) stratum() error {
	if len(e.role) != 0 {
		if err := e.header(system); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte(e.role)); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte(end_of_turn)); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) Write(s string) error {
	if e.seq%2 == 0 {
		return e.Prompt(s)
	}

	return e.Reply(s)
}

func (e *Encoder) Prompt(prompt string) error {
	if err := e.header(human); err != nil {
		return err
	}

	if _, err := e.w.Write([]byte(prompt)); err != nil {
		return err
	}

	if _, err := e.w.Write([]byte(end_of_turn)); err != nil {
		return err
	}

	// Note: requires as part of protocol
	if err := e.header(assistant); err != nil {
		return err
	}

	e.seq++
	return nil
}

func (e *Encoder) Reply(reply string) error {
	if _, err := e.w.Write([]byte(reply)); err != nil {
		return err
	}

	if _, err := e.w.Write([]byte(end_of_turn)); err != nil {
		return err
	}

	e.seq++
	return nil
}
