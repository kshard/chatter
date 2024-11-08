//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrock

import (
	"encoding"
	"encoding/json"
	"strings"

	"github.com/kshard/chatter"
)

// Amazon Titan Text model family
//
// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-titan-text.html
// See https://docs.aws.amazon.com/bedrock/latest/userguide/prompt-templates-and-examples.html
type Titan string

// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
const (
	TITAN_TEXT_LITE_V1    = Titan("amazon.titan-text-lite-v1")
	TITAN_TEXT_EXPRESS_V1 = Titan("amazon.titan-text-express-v1")
	TITAN_TEXT_PREMIER_V1 = Titan("amazon.titan-text-premier-v1:0")
)

func (v Titan) ID() string { return string(v) }

func (Titan) Encode(prompt encoding.TextMarshaler, opts *chatter.Options) ([]byte, error) {
	txt, err := prompt.MarshalText()
	if err != nil {
		return nil, err
	}

	req, err := json.Marshal(
		titanInquery{
			Prompt: string(txt),
			Config: titanInqueryConfig{
				Temperature: opts.Temperature,
				TopP:        opts.TopP,
				MaxTokens:   opts.Quota,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (Titan) Decode(data []byte) (r Reply, err error) {
	var reply titanChatter

	err = json.Unmarshal(data, &reply)
	if err != nil {
		return
	}

	r.UsedInputTokens = reply.UsedPromptTokens

	sb := strings.Builder{}
	for _, text := range reply.Result {
		sb.WriteString(strings.TrimPrefix(text.Text, "Bot:"))
		sb.WriteRune('\n')
		r.UsedReplyTokens += text.UsedTextTokens
	}
	r.Text = sb.String()

	return
}

type titanInquery struct {
	Prompt string             `json:"inputText"`
	Config titanInqueryConfig `json:"textGenerationConfig"`
}

type titanInqueryConfig struct {
	Temperature  float64  `json:"temperature,omitempty"`
	TopP         float64  `json:"topP,omitempty"`
	MaxTokens    int      `json:"maxTokenCount,omitempty"`
	StopSequence []string `json:"stopSequences,omitempty"`
}

type titanChatter struct {
	UsedPromptTokens int                `json:"inputTextTokenCount"`
	Result           []titanChatterText `json:"results"`
}

type titanChatterText struct {
	Text           string `json:"outputText"`
	UsedTextTokens int    `json:"tokenCount"`
	StopReason     string `json:"completionReason"`
}
