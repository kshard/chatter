//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import (
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func wireFactory() (provider.Encoder[*wireInput], error) {
	return &wireEncoder{
		req: &wireInput{
			Messages: make([]wireMessage, 0),
		},
	}, nil
}

type wireEncoder struct {
	req *wireInput
}

func (codec *wireEncoder) WithInferrer(inf provider.Inferrer) {
	cfg := wireInferenceConfig{}

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

func (codec *wireEncoder) WithCommand(cmd chatter.Cmd) {
	if codec.req.ToolConfig == nil {
		codec.req.ToolConfig = &wireToolConfig{
			ToolChoice: wireToolChoice{Auto: &struct{}{}},
			Tools:      []wireTool{},
		}
	}

	codec.req.ToolConfig.Tools = append(codec.req.ToolConfig.Tools, wireTool{
		ToolSpec: wireToolSpec{
			Name:        cmd.Cmd,
			Description: cmd.About,
			InputSchema: wireToolInputSchema{JSON: cmd.Schema},
		},
	})
}

func (codec *wireEncoder) AsStratum(stratum chatter.Stratum) error {
	codec.req.System = append(codec.req.System, wireContent{Text: string(stratum)})
	return nil
}

func (codec *wireEncoder) AsText(text chatter.Text) error {
	codec.req.Messages = append(codec.req.Messages, wireMessage{
		Role:    "user",
		Content: []wireContent{{Text: string(text)}},
	})
	return nil
}

func (codec *wireEncoder) AsPrompt(prompt *chatter.Prompt) error {
	codec.req.Messages = append(codec.req.Messages, wireMessage{
		Role:    "user",
		Content: []wireContent{{Text: prompt.String()}},
	})
	return nil
}

func (codec *wireEncoder) AsAnswer(answer *chatter.Answer) error {
	if len(answer.Yield) == 0 {
		return nil
	}

	content := make([]wireContent, 0, len(answer.Yield))
	for _, yield := range answer.Yield {
		content = append(content, wireContent{
			ToolResult: &wireToolResult{
				ToolUseID: yield.ID,
				Content:   []wireToolResultContent{{JSON: yield.Value}},
			},
		})
	}

	codec.req.Messages = append(codec.req.Messages, wireMessage{Role: "user", Content: content})
	return nil
}

func (codec *wireEncoder) AsReply(reply *chatter.Reply) error {
	content := make([]wireContent, 0, len(reply.Content))
	for _, block := range reply.Content {
		switch v := block.(type) {
		case chatter.Text:
			content = append(content, wireContent{Text: string(v)})
		case chatter.Invoke:
			content = append(content, wireContent{
				ToolUse: &wireToolUse{
					ToolUseID: v.Args.ID,
					Name:      v.Cmd,
					Input:     v.Args.Value,
				},
			})
		}
	}

	codec.req.Messages = append(codec.req.Messages, wireMessage{Role: "assistant", Content: content})
	return nil
}

func (codec *wireEncoder) Build() *wireInput { return codec.req }
