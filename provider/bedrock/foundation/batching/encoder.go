//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package batching

import (
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory() (provider.Encoder[*input], error) {
	return &encoder{
		req: &input{
			Messages: make([]message, 0),
		},
	}, nil
}

type encoder struct {
	req *input
}

func (codec *encoder) WithInferrer(inf provider.Inferrer) {
	cfg := inferenceConfig{}

	if inf.Temperature > 0.0 && inf.Temperature <= 1.0 {
		cfg.Temperature = inf.Temperature
	}
	if inf.TopP > 0.0 && inf.TopP <= 1.0 {
		cfg.TopP = inf.TopP
	}
	if inf.MaxTokens > 0 {
		cfg.MaxTokens = inf.MaxTokens
	}
	if inf.StopSequences != nil {
		cfg.StopSequences = inf.StopSequences
	}

	codec.req.InferenceConfig = &cfg
}

func (codec *encoder) WithCommand(cmd chatter.Cmd) {
	if codec.req.ToolConfig == nil {
		codec.req.ToolConfig = &toolConfig{
			ToolChoice: toolChoice{Auto: &struct{}{}},
			Tools:      []tool{},
		}
	}

	codec.req.ToolConfig.Tools = append(codec.req.ToolConfig.Tools, tool{
		ToolSpec: toolSpec{
			Name:        cmd.Cmd,
			Description: cmd.About,
			InputSchema: toolInputSchema{JSON: cmd.Schema},
		},
	})
}

func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	codec.req.System = append(codec.req.System, content{Text: string(stratum)})
	return nil
}

func (codec *encoder) AsText(text chatter.Text) error {
	codec.req.Messages = append(codec.req.Messages, message{
		Role:    "user",
		Content: []content{{Text: string(text)}},
	})
	return nil
}

func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	codec.req.Messages = append(codec.req.Messages, message{
		Role:    "user",
		Content: []content{{Text: prompt.String()}},
	})
	return nil
}

func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	if len(answer.Yield) == 0 {
		return nil
	}

	c := make([]content, 0, len(answer.Yield))
	for _, yield := range answer.Yield {
		c = append(c, content{
			ToolResult: &toolResult{
				ToolUseID: yield.ID,
				Content:   []toolResultContent{{JSON: yield.Value}},
			},
		})
	}

	codec.req.Messages = append(codec.req.Messages, message{Role: "user", Content: c})
	return nil
}

func (codec *encoder) AsReply(reply *chatter.Reply) error {
	c := make([]content, 0, len(reply.Content))
	for _, block := range reply.Content {
		switch v := block.(type) {
		case chatter.Text:
			c = append(c, content{Text: string(v)})
		case chatter.Invoke:
			c = append(c, content{
				ToolUse: &toolUse{
					ToolUseID: v.Args.ID,
					Name:      v.Cmd,
					Input:     v.Args.Value,
				},
			})
		}
	}

	codec.req.Messages = append(codec.req.Messages, message{Role: "assistant", Content: c})
	return nil
}

func (codec *encoder) Build() *input { return codec.req }
