package bedrock

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/fogfish/opts"
	"github.com/kshard/chatter/aio/provider"
)

//
// Configuration options for the Bedrock service
//

type Option = opts.Option[Client]

const defaultRegion = "us-west-2"

var (
	// Use aws.Config to config the client
	WithConfig = opts.FMap(optsFromConfig)

	// Use region for aws.Config
	WithRegion = opts.FMap(optsFromRegion)

	// Set AWS Bedrock Runtime
	WithRuntime = opts.ForType[Client, Runtime]()
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

// AWS Bedrock Runtime API
type Runtime interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

type Client struct{ api Runtime }

type Service[A, B any] struct {
	client    Client
	model     string
	undefined B
}

var _ provider.Service[int, int] = (*Service[int, int])(nil)

func New[A, B any](model string, opt ...Option) (*Service[A, B], error) {
	c := Client{}
	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}
	if c.api == nil {
		if err := optsFromRegion(&c, defaultRegion); err != nil {
			return nil, err
		}
	}

	return &Service[A, B]{
		client:    c,
		model:     model,
		undefined: *new(B),
	}, nil
}

func (s *Service[A, B]) Invoke(ctx context.Context, input A) (B, error) {
	req, err := json.Marshal(input)
	if err != nil {
		return s.undefined, err
	}

	inquiry := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(s.model),
		ContentType: aws.String("application/json"),
		Body:        req,
	}

	result, err := s.client.api.InvokeModel(ctx, inquiry)
	if err != nil {
		return s.undefined, err
	}

	var reply B

	err = json.Unmarshal(result.Body, &reply)
	if err != nil {
		return s.undefined, err
	}

	return reply, nil
}
