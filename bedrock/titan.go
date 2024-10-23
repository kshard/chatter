//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrock

import (
	"encoding/json"
	"strings"

	"github.com/kshard/chatter"
)

// Amazon Titan Text model family
//
// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-titan-text.html
// See https://docs.aws.amazon.com/bedrock/latest/userguide/prompt-templates-and-examples.html
type TitanText string

// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
const (
	TITAN_TEXT_LITE_V1    = TitanText("amazon.titan-text-lite-v1")
	TITAN_TEXT_EXPRESS_V1 = TitanText("amazon.titan-text-express-v1")
	TITAN_TEXT_PREMIER_V1 = TitanText("amazon.titan-text-premier-v1:0")
)

func (v TitanText) String() string { return string(v) }

func (TitanText) Formatter() chatter.Formatter {
	return chatter.NewFormatter("")
}

func (TitanText) Encode(c *Client, prompt *chatter.Prompt, opts *chatter.Options) ([]byte, error) {
	sb := strings.Builder{}
	c.formatter.ToString(&sb, prompt)

	inquery := titanInquery{
		Prompt: sb.String(),
		Config: titanInqueryConfig{
			Temperature: opts.Temperature,
			TopP:        opts.TopP,
			MaxTokens:   c.quotaTokensInReply,
		},
	}

	body, err := json.Marshal(inquery)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (TitanText) Decode(c *Client, data []byte) (string, error) {
	var reply titanChatter
	if err := json.Unmarshal(data, &reply); err != nil {
		return "", err
	}

	sb := strings.Builder{}

	c.consumedTokens += reply.UsedPromptTokens
	for _, text := range reply.Result {
		sb.WriteString(strings.TrimPrefix(text.Text, "Bot:"))
		sb.WriteRune('\n')
		c.consumedTokens += text.UsedTextTokens
	}

	return sb.String(), nil
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
