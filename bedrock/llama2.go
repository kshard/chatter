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

// Meta Llama2 model family
//
// See
// * https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html
// * https://replicate.com/blog/how-to-prompt-llama
type Llama2 string

// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
const (
	LLAMA2_13B_CHAT_V1 = Llama2("meta.llama2-13b-chat-v1")
	LLAMA2_70B_CHAT_V1 = Llama2("meta.llama2-70b-chat-v1")
)

func (v Llama2) String() string { return string(v) }

func (Llama2) Formatter() chatter.Formatter {
	return llama2Prompter{chatter.NewFormatter("")}
}

func (Llama2) Encode(c *Client, prompt *chatter.Prompt, opts *chatter.Options) ([]byte, error) {
	sb := strings.Builder{}
	c.formatter.ToString(&sb, prompt)

	inquery := llamaInquery{
		Prompt:      sb.String(),
		Temperature: opts.Temperature,
		TopP:        opts.TopP,
		MaxTokens:   c.quotaTokensInReply,
	}

	body, err := json.Marshal(inquery)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (Llama2) Decode(c *Client, data []byte) (string, error) {
	var reply llamaChatter
	if err := json.Unmarshal(data, &reply); err != nil {
		return "", err
	}

	c.consumedTokens += reply.UsedPromptTokens
	c.consumedTokens += reply.UsedTextTokens

	return reply.Text, nil
}

type llamaInquery struct {
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	MaxTokens   int     `json:"max_gen_len,omitempty"`
}

type llamaChatter struct {
	Text             string `json:"generation"`
	UsedPromptTokens int    `json:"prompt_token_count"`
	UsedTextTokens   int    `json:"generation_token_count"`
	StopReason       string `json:"stop_reason"`
}

type llama2Prompter struct{ chatter.Formatter }

func (p llama2Prompter) ToString(sb *strings.Builder, prompt *chatter.Prompt) {
	sb.WriteString("\n[INST]\n")
	p.Formatter.ToString(sb, prompt)
	sb.WriteString("\n[/INST]\n")
}
