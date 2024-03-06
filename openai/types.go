//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package openai

import (
	"os/user"
	"path/filepath"

	"github.com/fogfish/gurl/v2/http"
	ø "github.com/fogfish/gurl/v2/http/send"
	"github.com/jdxcode/netrc"
)

// Config option for the client
type Option func(*Client)

type ModelID string

const (
	GPT_35_TURBO_0125     = ModelID("gpt-3.5-turbo-0125")
	GPT_35_TURBO_INSTRUCT = ModelID("gpt-3.5-turbo-instruct")
	GPT_4                 = ModelID("gpt-4")
	GPT_4_32K             = ModelID("gpt-4-32k")
)

// Config the model
func WithModel(id ModelID) Option {
	return func(c *Client) {
		c.model = id
	}
}

// Config the http stack
func WithHTTP(opts ...http.Config) Option {
	return func(c *Client) {
		c.Stack = http.New(opts...)
	}
}

// Config the host, api.openai.com is default
func WithHost(host string) Option {
	return func(c *Client) {
		c.host = ø.Authority(host)
	}
}

// Config the secret explicitly
func WithSecret(secret string) Option {
	return func(c *Client) {
		c.secret = "Bearer " + secret
	}
}

// Config the secret from .netrc
func WithNetRC(host string) Option {
	return func(c *Client) {
		if c.secret != "" {
			return
		}

		usr, err := user.Current()
		if err != nil {
			panic(err)
		}

		n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
		if err != nil {
			panic(err)
		}

		machine := n.Machine(host)
		if machine == nil {
			return
			// panic(fmt.Errorf("undefined secret for host <%s> at ~/.netrc", host))
		}

		c.secret = "Bearer " + machine.Get("password")
	}
}

// OpenAI client
type Client struct {
	http.Stack
	host           ø.Authority
	secret         string
	model          ModelID
	consumedTokens int
}

// See https://platform.openai.com/docs/api-reference/chat/create
type modelInquery struct {
	Model       ModelID   `json:"model"`
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
