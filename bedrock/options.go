package bedrock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/fogfish/opts"
)

type Option = opts.Option[Client]

func (c *Client) checkRequired() error {
	return opts.Required(c,
		WithLLM(nil),
		WithBedrock(nil),
	)
}

var (
	// Set AWS Bedrock Foundational LLM
	//
	// This option is required.
	WithLLM = opts.ForType[Client, LLM]()

	// Use aws.Config to config the client
	WithConfig = opts.FMap(optsFromConfig)

	// Use region for aws.Config
	WithRegion = opts.FMap(optsFromRegion)

	// Set us-west-2 as default region
	WithDefaultRegion = WithRegion("us-east-1")

	// Set AWS Bedrock Runtime
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
	if c.api == nil {
		c.api = bedrockruntime.NewFromConfig(cfg)
	}

	return
}
