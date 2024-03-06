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
		WithNetRC(string(api.host)),
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
func (c *Client) Send(ctx context.Context, prompt *chatter.Prompt) (*chatter.Prompt, error) {
	seq := make([]message, 0)
	if prompt.Stratum != "" {
		seq = append(seq, message{Role: "system", Content: prompt.Stratum})
	}
	if prompt.Context != "" {
		seq = append(seq, message{Role: "system", Content: "Context: " + prompt.Context})
	}
	for _, m := range prompt.Messages {
		switch m.Role {
		case chatter.INQUIRY:
			seq = append(seq, message{Role: "user", Content: m.Content})
		case chatter.CHATTER:
			seq = append(seq, message{Role: "assistant", Content: m.Content})
		}
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
		return nil, err
	}

	c.consumedTokens += bag.Usage.UsedTokens
	if len(bag.Choices) == 0 {
		return nil, errors.New("empty response")
	}

	return prompt.Chatter(bag.Choices[0].Message.Content), nil
}
