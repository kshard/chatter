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
	"fmt"
	"strings"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/encoding/text"
)

// Amazon Titan Text model family
//
// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-titan-text.html
// See https://docs.aws.amazon.com/bedrock/latest/userguide/prompt-templates-and-examples.html
type Titan string

var _ chatter.LLM = Titan("")

// See https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html
const (
	TITAN_TEXT_LITE_V1    = Titan("amazon.titan-text-lite-v1")
	TITAN_TEXT_EXPRESS_V1 = Titan("amazon.titan-text-express-v1")
	TITAN_TEXT_PREMIER_V1 = Titan("amazon.titan-text-premier-v1:0")
)

func (v Titan) ModelID() string { return string(v) }

func (Titan) Encode(prompt []fmt.Stringer, opts ...chatter.Opt) (req []byte, err error) {
	if len(prompt) == 0 {
		err = fmt.Errorf("bad request, empty prompt")
		return
	}

	var ptext strings.Builder
	codec, err := text.NewEncoder(&ptext, "Bot: ", "User: ")
	if err != nil {
		return
	}

	for _, term := range prompt {
		switch v := term.(type) {
		case chatter.Stratum:
			err = codec.Stratum(string(v))
			if err != nil {
				return
			}
		case chatter.Reply:
			err = codec.Reply(v.Text)
			if err != nil {
				return
			}
		default:
			err = codec.Prompt(term.String())
			if err != nil {
				return
			}
		}
	}

	inquery := titanInquery{Prompt: ptext.String()}
	for _, opt := range opts {
		switch v := opt.(type) {
		case chatter.Temperature:
			inquery.Config.Temperature = float64(v)
		case chatter.TopP:
			inquery.Config.TopP = float64(v)
		case chatter.Quota:
			inquery.Config.MaxTokens = int(v)
		case chatter.StopSequence:
			inquery.Config.StopSequence = []string{string(v)}
		}
	}

	req, err = json.Marshal(inquery)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (Titan) Decode(data []byte) (r chatter.Reply, err error) {
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
