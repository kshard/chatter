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
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/kshard/chatter"
)

// AWS Bedrock client
type Client struct {
	api                *bedrockruntime.Client
	region             string
	model              Model
	formatter          chatter.Formatter
	quotaTokensInReply int
	consumedTokens     int
}

// Create client to AWS BedRock.
//
// By default `us-east-1` region is used, use config options to alter behavior.
func New(opts ...Option) (*Client, error) {
	client := &Client{region: "us-east-1"}

	for _, opt := range opts {
		opt(client)
	}

	if client.api == nil {
		api, err := newService(client.region)
		if err != nil {
			return nil, err
		}
		client.api = api
	}

	if client.model == nil {
		return nil, fmt.Errorf("undefined model")
	}

	return client, nil
}

func newService(region string) (*bedrockruntime.Client, error) {
	aws, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	return bedrockruntime.NewFromConfig(aws), nil
}

// Number of tokens consumed within the session
func (c *Client) ConsumedTokens() int { return c.consumedTokens }

// Prompt the model
func (c *Client) Prompt(ctx context.Context, prompt *chatter.Prompt, opts ...func(*chatter.Options)) (string, error) {
	opt := chatter.Options{Temperature: chatter.DefaultTemperature, TopP: chatter.DefaultTopP}
	for _, o := range opts {
		o(&opt)
	}

	body, err := c.model.Encode(c, prompt, &opt)
	if err != nil {
		return "", err
	}

	req := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(c.model.String()),
		ContentType: aws.String("application/json"),
		Accept:      aws.String("application/json"),
		Body:        body,
	}

	result, err := c.api.InvokeModel(ctx, req)
	if err != nil {
		return "", err
	}

	reply, err := c.model.Decode(c, result.Body)
	if err != nil {
		return "", err
	}

	return reply, nil
}
