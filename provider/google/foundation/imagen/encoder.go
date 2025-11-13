//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package imagen

import (
	"strings"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory(model string) func() (provider.Encoder[*input], error) {
	return func() (provider.Encoder[*input], error) {
		return &encoder{
			req: input{
				Model:  model,
				Prompt: strings.Builder{},
			},
		}, nil
	}
}

type encoder struct {
	req input
}

func (codec *encoder) WithInferrer(inf provider.Inferrer) {}

func (codec *encoder) WithCommand(cmd chatter.Cmd) {}

func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	codec.req.Prompt.WriteString(string(stratum))
	return nil
}

func (codec *encoder) AsText(text chatter.Text) error {
	codec.req.Prompt.WriteString(string(text))
	return nil
}

func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	codec.req.Prompt.WriteString(prompt.String())
	return nil
}

func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	// Not supported yet
	return nil
}

func (codec *encoder) AsReply(reply *chatter.Reply) error {
	codec.req.Prompt.WriteString(reply.String())
	return nil
}

func (codec *encoder) Build() *input {
	return &codec.req
}
