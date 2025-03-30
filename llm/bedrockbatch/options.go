//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrockbatch

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/fogfish/opts"
	"github.com/fogfish/stream"
	bedrockapi "github.com/kshard/chatter/bedrock"
)

type Option = opts.Option[Client]

func (c *Client) checkRequired() error {
	return opts.Required(c,
		WithRole(""),
		WithBucket(""),
		WithLLM(nil),
		WithBedrock(nil),
		WithFileSystem(nil),
	)
}

var (
	// The Amazon Resource Name (ARN) of the service role with permissions to carry
	// out and manage batch inference.
	//
	// This option is required.
	//
	// [See AWS documentation]: https://docs.aws.amazon.com/bedrock/latest/userguide/batch-iam-sr.html
	WithRole = opts.ForName[Client, string]("role")

	// Set AWS S3 bucket for input/output data
	//
	// This option is required.
	WithBucket = opts.ForName[Client, string]("bucket")

	// Set AWS Bedrock Foundational LLM
	//
	// This option is required.
	WithLLM = opts.ForType[Client, bedrockapi.LLM]()

	// Use aws.Config to config the client
	WithConfig = opts.FMap(optsFromConfig)

	// Use region for aws.Config
	WithRegion = opts.FMap(optsFromRegion)

	// Set us-west-2 as default region
	WithDefaultRegion = WithRegion("us-west-2")

	// Set file system client
	WithFileSystem = opts.ForType[Client, FileSystem]()

	// Set batch inference client
	WithBedrock = opts.ForType[Client, Bedrock]()
)

func optsFromRegion(c *Client, region string) error {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return err
	}

	return optsFromConfig(c, cfg)
}

func optsFromConfig(c *Client, cfg aws.Config) (err error) {
	if c.bedrock == nil {
		c.bedrock = bedrock.NewFromConfig(cfg)
	}

	if c.fsys == nil {
		c.fsys, err = stream.NewFS(c.bucket,
			stream.WithConfig(cfg),
			stream.WithIOTimeout(15*time.Minute),
		)
	}

	return
}
