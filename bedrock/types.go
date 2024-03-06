//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/embeddings
//

package bedrock

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/kshard/chatter"
)

// Config option for the client
type Option func(*Client)

// Config AWS endpoints
func WithConfig(cfg aws.Config) Option {
	return func(e *Client) {
		e.api = bedrockruntime.NewFromConfig(cfg)
	}
}

// Config bedrock model
func WithModel(model Model) Option {
	return func(c *Client) {
		c.model = model
	}
}

// Config tokens quota in reply
func WithQuotaTokensInReply(limit int) Option {
	return func(c *Client) {
		c.quotaTokensInReply = limit
	}
}

// Bedrock client
type Client struct {
	api                *bedrockruntime.Client
	model              Model
	quotaTokensInReply int
	consumedTokens     int
}

type Model interface {
	String() string
	encode(*Client, *chatter.Prompt) ([]byte, error)
	decode(*Client, []byte) (string, error)
}
