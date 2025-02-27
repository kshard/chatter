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
func (c *Client) Prompt(ctx context.Context, prompt []fmt.Stringer, opts ...chatter.Opt) (reply chatter.Reply, err error) {
	if len(prompt) == 0 {
		err = fmt.Errorf("bad request, empty prompt")
		return
	}

	seq := make([]message, 0)
	for _, term := range prompt {
		switch v := term.(type) {
		case chatter.Stratum:
			seq = append(seq, message{Role: "developer", Content: string(v)})
		case chatter.Reply:
			seq = append(seq, message{Role: "assistant", Content: term.String()})
		default:
			seq = append(seq, message{Role: "user", Content: term.String()})
		}
	}

	inquery := modelInquery{Model: c.llm, Messages: seq}
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

	bag, err := http.IO[modelChatter](c.WithContext(ctx),
		http.POST(
			ø.URI("%s/v1/chat/completions", c.host),
			ø.Accept.JSON,
			ø.Authorization.Set("Bearer "+c.secret),
			ø.ContentType.JSON,
			ø.Send(inquery),

			ƒ.Status.OK,
			ƒ.ContentType.JSON,
		),
	)
	if err != nil {
		return
	}

	c.usedInputTokens += bag.Usage.PromptTokens
	c.usedReplyTokens += bag.Usage.OutputTokens

	if len(bag.Choices) == 0 {
		err = errors.New("empty response")
		return
	}

	reply = chatter.Reply{
		Text:            bag.Choices[0].Message.Content,
		UsedInputTokens: bag.Usage.PromptTokens,
		UsedReplyTokens: bag.Usage.OutputTokens,
	}

	return
}
