package bedrock

import (
	"encoding"
	"encoding/json"
	"strings"

	"github.com/kshard/chatter"
)

// Meta Llama3 model family
//
// See
// * https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html
// * https://www.llama.com/docs/model-cards-and-prompt-formats/llama-guard-3
// * https://www.llama.com/docs/model-cards-and-prompt-formats/meta-llama-3/
type Llama3 string

var _ chatter.LLM = Llama2("")

// See model id
// https://docs.aws.amazon.com/bedrock/latest/userguide/model-ids.html#model-ids-arns
const (
	LLAMA3_0_8B_INSTRUCT   = Llama3("meta.llama3-8b-instruct-v1:0")
	LLAMA3_0_70B_INSTRUCT  = Llama3("meta.llama3-70b-instruct-v1:0")
	LLAMA3_1_8B_INSTRUCT   = Llama3("meta.llama3-1-8b-instruct-v1:0")
	LLAMA3_1_70B_INSTRUCT  = Llama3("meta.llama3-1-70b-instruct-v1:0")
	LLAMA3_1_405B_INSTRUCT = Llama3("meta.llama3-1-405b-instruct-v1:0")
	LLAMA3_2_1B_INSTRUCT   = Llama3("meta.llama3-2-1b-instruct-v1:0")
	LLAMA3_2_3B_INSTRUCT   = Llama3("meta.llama3-2-3b-instruct-v1:0")
	LLAMA3_2_11B_INSTRUCT  = Llama3("meta.llama3-2-11b-instruct-v1:0")
	LLAMA3_2_90B_INSTRUCT  = Llama3("meta.llama3-2-90b-instruct-v1:0")
)

func (v Llama3) ModelID() string { return string(v) }

func (v Llama3) Encode(prompt encoding.TextMarshaler, opts *chatter.Options) ([]byte, error) {
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

func (Llama3) encode(prompt []byte) string {
	var sb strings.Builder

	sb.WriteString("<|begin_of_text|>\n")
	sb.WriteString("<|start_header_id|>user<|end_header_id|>\n")
	sb.Write(prompt)
	sb.WriteString("<|eot_id|>\n")
	sb.WriteString("<|start_header_id|>assistant<|end_header_id|>\n")

	return sb.String()
}

func (Llama3) Decode(data []byte) (r chatter.Reply, err error) {
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
