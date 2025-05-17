package bedrock

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/encoding/llama3"
)

// Meta Llama3 model family
//
// See
// * https://docs.aws.amazon.com/bedrock/latest/userguide/model-parameters-meta.html
// * https://www.llama.com/docs/model-cards-and-prompt-formats/llama-guard-3
// * https://www.llama.com/docs/model-cards-and-prompt-formats/meta-llama-3/
type Llama3 string

var _ chatter.LLM = Llama3("")

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
	LLAMA3_3_70B_INSTRUCT  = Llama3("meta.llama3-3-70b-instruct-v1:0")
)

func (v Llama3) ModelID() string { return string(v) }

func (v Llama3) Encode(prompt []chatter.Message, opts ...chatter.Opt) (req []byte, err error) {
	if len(prompt) == 0 {
		err = fmt.Errorf("bad request, empty prompt")
		return
	}

	var ptext strings.Builder
	codec, err := llama3.NewEncoder(&ptext)
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
			err = codec.Reply(v.String())
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

	inquery := llamaInquery{Prompt: ptext.String()}
	for _, opt := range opts {
		switch v := opt.(type) {
		case chatter.Temperature:
			inquery.Temperature = float64(v)
		case chatter.TopP:
			inquery.TopP = float64(v)
		case chatter.Quota:
			inquery.MaxTokens = int(v)
		}
	}

	req, err = json.Marshal(inquery)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (Llama3) Decode(data []byte) (r chatter.Reply, err error) {
	var reply llamaChatter

	err = json.Unmarshal(data, &reply)
	if err != nil {
		return
	}

	r.Content = []chatter.Content{
		chatter.Text(reply.Text),
	}
	r.Usage.InputTokens = reply.UsedPromptTokens
	r.Usage.ReplyTokens = reply.UsedTextTokens

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
