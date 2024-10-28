package bedrock

import (
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

func (v Llama3) String() string { return string(v) }

func (Llama3) Formatter() chatter.Formatter {
	return llama3Prompter{chatter.NewFormatter("")}
}

func (Llama3) Encode(c *Client, prompt *chatter.Prompt, opts *chatter.Options) ([]byte, error) {
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

func (Llama3) Decode(c *Client, data []byte) (string, error) {
	var reply llamaChatter
	if err := json.Unmarshal(data, &reply); err != nil {
		return "", err
	}

	c.consumedTokens += reply.UsedPromptTokens
	c.consumedTokens += reply.UsedTextTokens

	return reply.Text, nil
}

type llama3Prompter struct{ chatter.Formatter }

func (p llama3Prompter) ToString(sb *strings.Builder, prompt *chatter.Prompt) {
	sb.WriteString("<|begin_of_text|>\n")
	sb.WriteString("<|start_header_id|>user<|end_header_id|>\n")
	p.Formatter.ToString(sb, prompt)
	sb.WriteString("<|eot_id|>\n")
	sb.WriteString("<|start_header_id|>assistant<|end_header_id|>\n")
}
