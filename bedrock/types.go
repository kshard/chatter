//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
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

// Config AWS region
func WithRegion(region string) Option {
	return func(c *Client) {
		c.region = region
	}
}

// Config Bedrock model
func WithModel(model Model) Option {
	return func(c *Client) {
		c.model = model
		c.formatter = model.Formatter()
	}
}

// Config Formatter
func WithFormatter(formatter chatter.Formatter) Option {
	return func(c *Client) {
		c.formatter = formatter
	}
}

// Config tokens quota in reply
func WithQuotaTokensInReply(quota int) Option {
	return func(c *Client) {
		c.quotaTokensInReply = quota
	}
}

type Model interface {
	String() string
	Formatter() chatter.Formatter
	Encode(*Client, *chatter.Prompt, *chatter.Options) ([]byte, error)
	Decode(*Client, []byte) (string, error)
}
