//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package openai

import (
	"context"
	"errors"
	"strings"

	"github.com/kshard/chatter"

	"github.com/fogfish/gurl/v2/http"
	ƒ "github.com/fogfish/gurl/v2/http/recv"
	ø "github.com/fogfish/gurl/v2/http/send"
)

// Creates OpenAI Chat (completion) client.
//
// By default OpenAI reads access token from `~/.netrc`,
// supply custom secret `WithSecret(secret string)` if needed.
//
// The client is configurable using
//
//	WithSecret(secret string)
//	WithNetRC(host string)
//	WithModel(...)
//	WithHTTP(opts ...http.Config)
func New(opts ...Option) (*Client, error) {
	api := &Client{host: ø.Authority("https://api.openai.com")}

	defs := []Option{
		WithModel(GPT_35_TURBO_0125),
		WithNetRC(string("api.openai.com")),
	}

	for _, opt := range defs {
		opt(api)
	}

	for _, opt := range opts {
		opt(api)
	}

	if api.Stack == nil {
		api.Stack = http.New()
	}

	return api, nil
}

// Number of tokens consumed within the session
func (c *Client) ConsumedTokens() int { return c.consumedTokens }

// Send prompt
func (c *Client) Prompt(ctx context.Context, prompt *chatter.Prompt, opts ...func(*chatter.Options)) (string, error) {
	var sb strings.Builder
	c.formatter.ToString(&sb, prompt)

	seq := []message{
		{Role: "user", Content: sb.String()},
	}

	bag, err := http.IO[modelChatter](c.WithContext(ctx),
		http.POST(
			ø.URI("%s/v1/chat/completions", c.host),
			ø.Accept.JSON,
			ø.Authorization.Set(c.secret),
			ø.ContentType.JSON,
			ø.Send(modelInquery{
				Model:     c.model,
				Messages:  seq,
				MaxTokens: c.quotaTokensInReply,
			}),

			ƒ.Status.OK,
			ƒ.ContentType.JSON,
		),
	)
	if err != nil {
		return "", err
	}

	c.consumedTokens += bag.Usage.UsedTokens
	if len(bag.Choices) == 0 {
		return "", errors.New("empty response")
	}

	return bag.Choices[0].Message.Content, nil
}
