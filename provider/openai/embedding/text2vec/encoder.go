//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text2vec

import (
	"strings"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory(model string, dimensions int) func() (provider.Encoder[*input], error) {
	return func() (provider.Encoder[*input], error) {
		codec := &encoder{
			w: strings.Builder{},
			req: input{
				Model:      model,
				Dimensions: dimensions,
			},
		}
		return codec, nil
	}
}

func (codec *encoder) WithInferrer(inferrer provider.Inferrer) {}
func (codec *encoder) WithCommand(cmd chatter.Cmd)             {}

func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	return nil
}

func (codec *encoder) AsText(text chatter.Text) error {
	codec.w.WriteString(string(text))
	return nil
}

func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	codec.w.WriteString(prompt.String())
	return nil
}

func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	return nil
}

func (codec *encoder) AsReply(reply *chatter.Reply) error {
	return nil
}

func (codec *encoder) Build() *input {
	codec.req.Text = codec.w.String()
	return &codec.req
}
