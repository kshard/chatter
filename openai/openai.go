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
	"fmt"

	"github.com/kshard/chatter"

	"github.com/fogfish/gurl/v2/http"
	ƒ "github.com/fogfish/gurl/v2/http/recv"
	ø "github.com/fogfish/gurl/v2/http/send"
	"github.com/fogfish/opts"
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
func New(opt ...Option) (*Client, error) {
	api := Client{host: ø.Authority("https://api.openai.com")}

	if err := opts.Apply(&api, opt); err != nil {
		return nil, err
	}

	if api.Stack == nil {
		api.Stack = http.New()
	}

	return &api, api.checkRequired()
}

func (c *Client) UsedInputTokens() int { return c.usedInputTokens }
func (c *Client) UsedReplyTokens() int { return c.usedReplyTokens }

// Send prompt
func (c *Client) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...func(*chatter.Options)) (chatter.Text, error) {
	if len(prompt) == 0 {
		return "", fmt.Errorf("bad request, empty prompt")
	}

	opt := chatter.NewOptions()
	for _, o := range opts {
		o(&opt)
	}

	seq := make([]message, 0)
	switch v := prompt[0].(type) {
	case chatter.Stratum:
		seq = append(seq, message{Role: "developer", Content: string(v)})
		prompt = prompt[1:]
	}

	for i, msg := range prompt {
		if i%2 == 0 {
			seq = append(seq, message{Role: "user", Content: msg.String()})
		} else {
			seq = append(seq, message{Role: "assistant", Content: msg.String()})
		}
	}

	bag, err := http.IO[modelChatter](c.WithContext(ctx),
		http.POST(
			ø.URI("%s/v1/chat/completions", c.host),
			ø.Accept.JSON,
			ø.Authorization.Set("Bearer "+c.secret),
			ø.ContentType.JSON,
			ø.Send(modelInquery{
				Model:       c.llm,
				Messages:    seq,
				MaxTokens:   opt.Quota,
				Temperature: opt.Temperature,
				TopP:        opt.TopP,
			}),

			ƒ.Status.OK,
			ƒ.ContentType.JSON,
		),
	)
	if err != nil {
		return "", err
	}

	c.usedInputTokens += bag.Usage.PromptTokens
	c.usedReplyTokens += bag.Usage.OutputTokens

	if len(bag.Choices) == 0 {
		return "", errors.New("empty response")
	}

	return chatter.Text(bag.Choices[0].Message.Content), nil
}
