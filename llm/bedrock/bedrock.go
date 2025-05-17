//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/fogfish/opts"
	"github.com/kshard/chatter"
)

// AWS Bedrock client
type Client struct {
	api   Bedrock
	llm   LLM
	usage chatter.Usage
}

var _ chatter.Chatter = (*Client)(nil)

// Create client to AWS BedRock.
//
// By default `us-east-1` region is used, use config options to alter behavior.
func New(opt ...Option) (*Client, error) {
	c := Client{}

	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}

	if c.api == nil {
		if err := optsFromRegion(&c, defaultRegion); err != nil {
			return nil, err
		}
	}

	return &c, c.checkRequired()
}

func (c *Client) Usage() chatter.Usage { return c.usage }

// Prompt the model
func (c *Client) Prompt(ctx context.Context, prompt []chatter.Message, opts ...chatter.Opt) (*chatter.Reply, error) {
	req, err := c.llm.Encode(prompt, opts...)
	if err != nil {
		return nil, err
	}

	inquiry := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.llm.ModelID()),
		ContentType: aws.String("application/json"),
		Body:        req,
	}

	result, err := c.api.InvokeModel(ctx, inquiry)
	if err != nil {
		return nil, err
	}

	reply, err := c.llm.Decode(result.Body)
	if err != nil {
		return nil, err
	}

	c.usage.InputTokens += reply.Usage.InputTokens
	c.usage.ReplyTokens += reply.Usage.ReplyTokens

	return &reply, nil
}
