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

// Meta Llama2 model family
//
// See
// * https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html
// * https://replicate.com/blog/how-to-prompt-llama
type Llama2 string

var _ chatter.LLM = Llama2("")

// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
const (
	LLAMA2_13B_CHAT_V1 = Llama2("meta.llama2-13b-chat-v1")
	LLAMA2_70B_CHAT_V1 = Llama2("meta.llama2-70b-chat-v1")
)

func (v Llama2) ModelID() string { return string(v) }

func (v Llama2) Encode(prompt encoding.TextMarshaler, opts *chatter.Options) ([]byte, error) {
	txt, err := prompt.MarshalText()
	if err != nil {
		return nil, err
	}

	req, err := json.Marshal(
		llamaInquery{
			Prompt:      v.encode(txt),
			Temperature: opts.Temperature,
			TopP:        opts.TopP,
			MaxTokens:   opts.Quota,
		},
	)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (Llama2) encode(prompt []byte) string {
	var sb strings.Builder

	sb.WriteString("\n[INST]\n")
	sb.Write(prompt)
	sb.WriteString("\n[/INST]\n")

	return sb.String()
}

func (Llama2) Decode(data []byte) (r chatter.Reply, err error) {
	var reply llamaChatter

	err = json.Unmarshal(data, &reply)
	if err != nil {
		return
	}

	r.Text = reply.Text
	r.UsedInputTokens = reply.UsedPromptTokens
	r.UsedReplyTokens = reply.UsedTextTokens

	return
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
