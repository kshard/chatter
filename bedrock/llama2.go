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

// Meta Llama 2 Chat model
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

func (Llama2) encode(c *Client, prompt *chatter.Prompt) ([]byte, error) {
	ctx := strings.Builder{}
	if prompt.Stratum != "" {
		ctx.WriteString(prompt.Stratum)
		ctx.WriteRune('\n')
	}
	if prompt.Context != "" {
		ctx.WriteString("Context: ")
		ctx.WriteString(prompt.Context)
		ctx.WriteRune('\n')
	}

	req := strings.Builder{}

	for i := 0; i < len(prompt.Messages)-1; i++ {
		msg := prompt.Messages[i]

		switch msg.Role {
		case chatter.INQUIRY:
			req.WriteString("[INST]\n")
			req.WriteString(msg.Content)
			req.WriteString("\n[/INST]\n")
		case chatter.CHATTER:
			req.WriteString(msg.Content)
			req.WriteRune('\n')
		}
	}

	tail := prompt.Messages[len(prompt.Messages)-1]
	if tail.Role == chatter.INQUIRY {
		req.WriteString("[INST]\n")
		if ctx.Len() > 0 {
			req.WriteString(ctx.String())
			req.WriteRune('\n')
		}
		req.WriteString(tail.Content)
		req.WriteString("\n[/INST]\n")
	}

	inquery := llamaInquery{
		Prompt:    req.String(),
		MaxTokens: c.quotaTokensInReply,
	}

	body, err := json.Marshal(inquery)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (Llama2) decode(c *Client, data []byte) (string, error) {
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
