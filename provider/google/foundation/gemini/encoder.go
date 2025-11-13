//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package gemini

import (
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
	"google.golang.org/genai"
)

type encoder struct {
	req input
}

type input struct {
	Model  string
	Prompt []*genai.Content
	Params genai.GenerateContentConfig
}

func factory(model string) func() (provider.Encoder[*input], error) {
	return func() (provider.Encoder[*input], error) {
		return &encoder{
			req: input{
				Model:  model,
				Prompt: make([]*genai.Content, 0),
			},
		}, nil
	}
}

func (codec *encoder) WithInferrer(inf provider.Inferrer) {
	if inf.Temperature > 0.0 && inf.Temperature <= 1.0 {
		f := float32(inf.Temperature)
		codec.req.Params.Temperature = &f
	}

	if inf.TopP > 0.0 && inf.TopP <= 1.0 {
		f := float32(inf.TopP)
		codec.req.Params.TopP = &f
	}

	if inf.MaxTokens > 0 {
		codec.req.Params.MaxOutputTokens = int32(inf.MaxTokens)
	}

	if inf.StopSequences != nil {
		codec.req.Params.StopSequences = inf.StopSequences
	}
}

func (codec *encoder) WithCommand(cmd chatter.Cmd) {
	// Not supported yet
}

func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	codec.req.Prompt = append(codec.req.Prompt,
		&genai.Content{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{{Text: string(stratum)}},
		},
	)
	return nil
}

func (codec *encoder) AsText(text chatter.Text) error {
	codec.req.Prompt = append(codec.req.Prompt,
		&genai.Content{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{{Text: string(text)}},
		},
	)
	return nil
}

func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	codec.req.Prompt = append(codec.req.Prompt,
		&genai.Content{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{{Text: prompt.String()}},
		},
	)
	return nil
}

func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	// Not supported yet
	return nil
}

func (codec *encoder) AsReply(reply *chatter.Reply) error {
	codec.req.Prompt = append(codec.req.Prompt,
		&genai.Content{
			Role:  genai.RoleModel,
			Parts: []*genai.Part{{Text: reply.String()}},
		},
	)
	return nil
}

func (codec *encoder) Build() *input {
	return &codec.req
}
