//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package nova

import (
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory() (provider.Encoder[*input], error) {
	codec := &encoder{
		req: input{
			Messages: []message{},
		},
	}
	return codec, nil
}

func (codec *encoder) WithInferrer(inf provider.Inferrer) {
	if inf.Temperature > 0.0 && inf.Temperature <= 1.0 {
		if codec.req.InferenceConfig == nil {
			codec.req.InferenceConfig = &inferenceConfig{}
		}
		codec.req.InferenceConfig.Temperature = inf.Temperature
	}
	if inf.TopP > 0.0 && inf.TopP <= 1.0 {
		if codec.req.InferenceConfig == nil {
			codec.req.InferenceConfig = &inferenceConfig{}
		}
		codec.req.InferenceConfig.TopP = inf.TopP
	}
	if inf.TopK > 0 {
		if codec.req.InferenceConfig == nil {
			codec.req.InferenceConfig = &inferenceConfig{}
		}
		codec.req.InferenceConfig.TopK = int(inf.TopK)
	}
	if inf.MaxTokens > 0 {
		if codec.req.InferenceConfig == nil {
			codec.req.InferenceConfig = &inferenceConfig{}
		}
		codec.req.InferenceConfig.MaxTokens = inf.MaxTokens
	}
	if inf.StopSequences != nil {
		if codec.req.InferenceConfig == nil {
			codec.req.InferenceConfig = &inferenceConfig{}
		}
		codec.req.InferenceConfig.StopSequences = inf.StopSequences
	}
}

func (codec *encoder) WithCommand(cmd chatter.Cmd) {
	// Nova doesn't support tools in this bedrock implementation
}

// AsStratum processes a Stratum message (system role)
func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	msg := content{Text: string(stratum)}
	codec.req.System = append(codec.req.System, msg)
	return nil
}

// AsText processes a Text message as user input
func (codec *encoder) AsText(text chatter.Text) error {
	txt := content{Text: string(text)}
	msg := message{Role: "user", Content: []content{txt}}
	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

// AsPrompt processes a Prompt message by converting it to string
func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	txt := content{Text: prompt.String()}
	msg := message{Role: "user", Content: []content{txt}}
	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

// AsAnswer processes an Answer message (tool results)
func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	if len(answer.Yield) == 0 {
		return nil
	}

	msg := message{
		Role:    "user",
		Content: []content{},
	}

	for _, yield := range answer.Yield {
		txt := content{Text: string(yield.Value)}
		msg.Content = append(msg.Content, txt)
	}

	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

// AsReply processes a Reply message (assistant response)
func (codec *encoder) AsReply(reply *chatter.Reply) error {
	txt := content{Text: reply.String()}
	msg := message{Role: "assistant", Content: []content{txt}}
	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

func (codec *encoder) Build() *input {
	return &codec.req
}
