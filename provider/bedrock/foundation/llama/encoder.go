//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llama

import (
	"strings"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory() (provider.Encoder[*input], error) {
	codec := &encoder{
		w:   strings.Builder{},
		req: input{},
	}
	codec.w.WriteString(begin_of_text)
	return codec, nil
}

func (codec *encoder) writeHeader(actor string) error {
	codec.w.WriteString(start_header_id)
	codec.w.WriteString(actor)
	codec.w.WriteString(end_header_id)
	return nil
}

func (codec *encoder) WithInferrer(inferrer provider.Inferrer) {
	codec.req.Temperature = inferrer.Temperature
	codec.req.TopP = inferrer.TopP
	codec.req.MaxTokens = inferrer.MaxTokens
}

func (codec *encoder) WithCommand(cmd chatter.Cmd) {
	// Llama3 doesn't support tools in this bedrock implementation
}

// AsStratum processes a Stratum message (system role)
func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	codec.writeHeader(system)
	codec.w.WriteString(string(stratum))
	codec.w.WriteString(end_of_turn)
	return nil
}

// AsText processes a Text message as user input
func (codec *encoder) AsText(text chatter.Text) error {
	codec.writeHeader(human)
	codec.w.WriteString(string(text))
	codec.w.WriteString(end_of_turn)
	// Note: required as part of Llama3 protocol
	codec.writeHeader(assistant)
	return nil
}

// AsPrompt processes a Prompt message by converting it to string
func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	codec.writeHeader(human)
	codec.w.WriteString(prompt.String())
	codec.w.WriteString(end_of_turn)
	// Note: required as part of Llama3 protocol
	codec.writeHeader(assistant)
	return nil
}

// AsAnswer processes an Answer message (tool results)
func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	if len(answer.Yield) == 0 {
		return nil
	}

	codec.writeHeader(human)
	for _, content := range answer.Yield {
		codec.w.Write(content.Value)
	}
	codec.w.WriteString(end_of_turn)
	// Note: required as part of Llama3 protocol
	codec.writeHeader(assistant)

	return nil
}

// AsReply processes a Reply message (assistant response)
func (codec *encoder) AsReply(reply *chatter.Reply) error {
	codec.w.WriteString(reply.String())
	codec.w.WriteString(end_of_turn)
	return nil
}

func (codec *encoder) Build() *input {
	codec.req.Prompt = codec.w.String()
	return &codec.req
}
