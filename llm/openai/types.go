//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package openai

import (
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/fogfish/gurl/v2/http"
	ø "github.com/fogfish/gurl/v2/http/send"
	"github.com/fogfish/opts"
	"github.com/jdxcode/netrc"
	"github.com/kshard/chatter"
)

type LLM string

const (
	GPT_35_TURBO_0125     = LLM("gpt-3.5-turbo-0125")
	GPT_35_TURBO_INSTRUCT = LLM("gpt-3.5-turbo-instruct")
	GPT_4                 = LLM("gpt-4")
	GPT_4_32K             = LLM("gpt-4-32k")
	GPT_4O                = LLM("gpt-4o")
	GPT_4O_MINI           = LLM("gpt-4o-mini")
	GPT_O1                = LLM("o1")
	GPT_O3_MINI           = LLM("o3-mini")
)

type Option = opts.Option[Client]

func (c *Client) checkRequired() error {
	return opts.Required(c,
		WithLLM(""),
	)
}

var (
	// Set OpenAI LLM
	//
	// This option is required.
	WithLLM = opts.ForType[Client, LLM]()

	// Config HTTP stack
	WithHTTP = opts.Use[Client](http.NewStack)

	// Config the host, api.openai.com is default
	WithHost = opts.ForType[Client, ø.Authority]()

	// Config API secret key
	WithSecret = opts.ForName[Client, string]("secret")

	// Set api secret from ~/.netrc file
	WithNetRC = opts.FMap(withNetRC)

	// Config the system role, the default one is `system`.
	// The role has been switch to `developer` on newer OpenAI LLMs
	WithRoleSystem = opts.ForName[Client, string]("roleSystem")
)

func withNetRC(h *Client, host string) error {
	if h.secret != "" {
		return nil
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		return err
	}

	machine := n.Machine(host)
	if machine == nil {
		return fmt.Errorf("undefined secret for host <%s> at ~/.netrc", host)
	}

	h.secret = machine.Get("password")
	return nil
}

// OpenAI client
type Client struct {
	http.Stack
	host       ø.Authority
	secret     string
	roleSystem string
	llm        LLM
	usage      chatter.Usage
}

// See https://platform.openai.com/docs/api-reference/chat/create
type modelInquery struct {
	Model       LLM       `json:"model"`
	Messages    []message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type modelChatter struct {
	ID      string   `json:"id"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Message message `json:"message"`
}

type usage struct {
	PromptTokens int `json:"prompt_tokens"`
	OutputTokens int `json:"completion_tokens"`
	UsedTokens   int `json:"total_tokens"`
}
