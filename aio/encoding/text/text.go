//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text

import "io"

type Encoder struct {
	w         io.Writer
	assistant string
	human     string
}

func NewEncoder(w io.Writer, assistant, human string) (*Encoder, error) {
	codec := &Encoder{
		w:         w,
		assistant: assistant,
		human:     human,
	}

	return codec, nil
}

func (e *Encoder) Stratum(statum string) error {
	if _, err := e.w.Write([]byte(statum)); err != nil {
		return err
	}
	if _, err := e.w.Write([]byte(".\n\n")); err != nil {
		return err
	}

	return nil
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

	return nil
}

func (e *Encoder) Reply(reply string) error {
	if _, err := e.w.Write([]byte(reply)); err != nil {
		return err
	}

	if _, err := e.w.Write([]byte("\n\n")); err != nil {
		return err
	}

	return nil
}
