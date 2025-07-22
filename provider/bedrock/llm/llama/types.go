package llama

import (
	"strings"

	"github.com/kshard/chatter/aio/provider"
	"github.com/kshard/chatter/provider/bedrock"
)

const (
	begin_of_text   = "<|begin_of_text|>"
	start_header_id = "\n<|start_header_id|>"
	end_header_id   = "<|end_header_id|>\n"
	end_of_turn     = "\n<|eot_id|>\n"
	system          = "system"
	assistant       = "assistant"
	human           = "user"
)

type input struct {
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	MaxTokens   int     `json:"max_gen_len,omitempty"`
}

type reply struct {
	Text             string `json:"generation"`
	UsedPromptTokens int    `json:"prompt_token_count"`
	UsedTextTokens   int    `json:"generation_token_count"`
	StopReason       string `json:"stop_reason"`
}

type encoder struct {
	w   strings.Builder
	req input
}

type decoder struct{}

type Llama = provider.Provider[*input, *reply]

func New(model string, opts ...bedrock.Option) (*Llama, error) {
	service, err := bedrock.New[*input, *reply](model, opts...)
	if err != nil {
		return nil, err
	}

	return provider.New(factory, decoder{}, service), nil
}
