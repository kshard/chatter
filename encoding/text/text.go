//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text

import "io"

// const (
// 	you_are   = "You are "
// 	assistant = "Bot: "
// 	human     = "User: "
// )

type Encoder struct {
	w         io.Writer
	assistant string
	human     string
	role      string
	seq       int
}

func NewEncoder(w io.Writer, assistant, human, role string) (*Encoder, error) {
	codec := &Encoder{
		w:         w,
		assistant: assistant,
		human:     human,
		role:      role,
		seq:       0,
	}
	if err := codec.session(); err != nil {
		return nil, err
	}

	return codec, nil
}

func (e *Encoder) session() error {
	if err := e.stratum(); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) stratum() error {
	if len(e.role) != 0 {
		if _, err := e.w.Write([]byte(e.role)); err != nil {
			return err
		}
		if _, err := e.w.Write([]byte(".\n\n")); err != nil {
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
	if _, err := e.w.Write([]byte(e.human)); err != nil {
		return err
	}

	if _, err := e.w.Write([]byte(prompt)); err != nil {
		return err
	}

	if _, err := e.w.Write([]byte("\n\n")); err != nil {
		return err
	}

	// Note: requires as part of protocol
	if _, err := e.w.Write([]byte(e.assistant)); err != nil {
		return err
	}

	e.seq++
	return nil
}

func (e *Encoder) Reply(reply string) error {
	if _, err := e.w.Write([]byte(reply)); err != nil {
		return err
	}

	if _, err := e.w.Write([]byte("\n\n")); err != nil {
		return err
	}

	e.seq++
	return nil
}
