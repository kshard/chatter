package converse

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/fogfish/opts"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

//
// Configuration options for the Bedrock Converse
//

type Option = opts.Option[Service]

const defaultRegion = "us-west-2"

var (
	// Use aws.Config to config the client
	WithConfig = opts.FMap(optsFromConfig)

	// Use region for aws.Config
	WithRegion = opts.FMap(optsFromRegion)

	// Set AWS Bedrock Runtime
	WithRuntime = opts.ForType[Service, Runtime]()

	// Set command-line registry
	WithRegistry = opts.ForType[Service, chatter.Registry]()
)

func optsFromRegion(c *Service, region string) error {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRegion(region),
	)
	if err != nil {
		return err
	}

	return optsFromConfig(c, cfg)
}

func optsFromConfig(c *Service, cfg aws.Config) (err error) {
	if c.api == nil {
		c.api = bedrockruntime.NewFromConfig(cfg)
	}

	return
}

type Runtime interface {
	Converse(ctx context.Context, params *bedrockruntime.ConverseInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.ConverseOutput, error)
}

type Service struct {
	api      Runtime
	registry chatter.Registry
}

var _ provider.Service[*bedrockruntime.ConverseInput, *bedrockruntime.ConverseOutput] = (*Service)(nil)

func (s *Service) Invoke(ctx context.Context, input *bedrockruntime.ConverseInput) (*bedrockruntime.ConverseOutput, error) {
	return s.api.Converse(ctx, input)
}

type Converse = provider.Provider[*bedrockruntime.ConverseInput, *bedrockruntime.ConverseOutput]

func New(model string, opt ...Option) (*Converse, error) {
	c := &Service{}

	if err := opts.Apply(c, opt); err != nil {
		return nil, err
	}

	if c.api == nil {
		if err := optsFromRegion(c, defaultRegion); err != nil {
			return nil, err
		}
	}

	return provider.New(factory(model, c.registry), decoder{}, c), nil
}
