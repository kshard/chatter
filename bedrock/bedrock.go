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
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/kshard/chatter"
)

// Creates AWS BedRock embeddings client.
//
// By default `us-east-1` region is used, supply custom `aws.Config`
// to alter behavior.
//
// The client is configurable using
//
//	WithConfig(cfg aws.Config)
//	WithModel(model Model)
//	WithQuotaTokensInReply(quota int)
func New(opts ...Option) (*Client, error) {
	client := &Client{}

	defs := []Option{
		WithModel(TITAN_TEXT_LITE_V1),
	}

	for _, opt := range defs {
		opt(client)
	}

	for _, opt := range opts {
		opt(client)
	}

	api, err := newService(client)
	if err != nil {
		return nil, err
	}
	client.api = api

	return client, nil
}

func newService(client *Client) (*bedrockruntime.Client, error) {
	if client.api != nil {
		return client.api, nil
	}

	aws, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return nil, err
	}

	return bedrockruntime.NewFromConfig(aws), nil
}

// Number of tokens consumed within the session
func (c *Client) ConsumedTokens() int { return c.consumedTokens }

// Calculates embedding vector
func (c *Client) Send(ctx context.Context, prompt *chatter.Prompt) (*chatter.Prompt, error) {
	body, err := c.model.encode(c, prompt)
	if err != nil {
		return nil, err
	}

	req := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.model.String()),
		ContentType: aws.String("application/json"),
		Body:        body,
	}

	result, err := c.api.InvokeModel(ctx, req)
	if err != nil {
		return nil, err
	}

	reply, err := c.model.decode(c, result.Body)
	if err != nil {
		return nil, err
	}

	return prompt.Chatter(reply), nil
}
