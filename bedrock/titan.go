//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/embeddings
//

package bedrock

import (
	"encoding/json"
	"strings"

	"github.com/kshard/chatter"
)

// Amazon Titan Text model
// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-titan-text.html
// See https://docs.aws.amazon.com/bedrock/latest/userguide/prompt-templates-and-examples.html
type TitanText string

// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
const (
	TITAN_TEXT_LITE_V1    = TitanText("amazon.titan-text-lite-v1")
	TITAN_TEXT_EXPRESS_V1 = TitanText("amazon.titan-text-express-v1")
)

func (v TitanText) String() string { return string(v) }

func (TitanText) encode(c *Client, prompt *chatter.Prompt) ([]byte, error) {
	ctx := strings.Builder{}
	if prompt.Stratum != "" {
		ctx.WriteString(prompt.Stratum)
		ctx.WriteRune(' ')
	}
	if prompt.Context != "" {
		ctx.WriteString("The context for the input is \"")
		ctx.WriteString(prompt.Context)
		ctx.WriteString("\".")
	}

	req := strings.Builder{}

	for i := 0; i < len(prompt.Messages)-1; i++ {
		msg := prompt.Messages[i]

		switch msg.Role {
		case chatter.INQUIRY:
			req.WriteString("User: ")
			req.WriteString(msg.Content)
			req.WriteRune('\n')
		case chatter.CHATTER:
			req.WriteString("Bot: ")
			req.WriteString(msg.Content)
			req.WriteRune('\n')
		}
	}

	tail := prompt.Messages[len(prompt.Messages)-1]
	if tail.Role == chatter.INQUIRY {
		req.WriteString("User: ")
		if ctx.Len() > 0 {
			req.WriteString(ctx.String())
			req.WriteRune(' ')
		}
		req.WriteString(tail.Content)
		req.WriteRune('\n')
	}

	inquery := titanInquery{
		Prompt: req.String(),
		Config: titanInqueryConfig{
			MaxTokens: c.quotaTokensInReply,
		},
	}

	body, err := json.Marshal(inquery)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (TitanText) decode(c *Client, data []byte) (string, error) {
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
