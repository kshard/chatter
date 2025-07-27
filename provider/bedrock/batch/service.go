//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package batch

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/opts"
	"github.com/fogfish/stream"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

//
// Configuration options for Bedrock
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
		c.api = bedrock.NewFromConfig(cfg)
	}

	return
}

type Runtime interface {
	CreateModelInvocationJob(ctx context.Context, params *bedrock.CreateModelInvocationJobInput, optFns ...func(*bedrock.Options)) (*bedrock.CreateModelInvocationJobOutput, error)
}

type Client struct {
	api   Runtime
	fs    *FileSystem
	model string
}

type Provider[A, B any] struct {
	factory provider.Factory[A]
	decoder provider.Decoder[B]
	client  Client
}

func New[A, B any](
	fs *FileSystem,
	model string,
	factory provider.Factory[A],
	decoder provider.Decoder[B],
	opt ...Option,
) (*Provider[A, B], error) {
	c := Client{
		fs:    fs,
		model: model,
	}

	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}

	if c.api == nil {
		if err := optsFromRegion(&c, defaultRegion); err != nil {
			return nil, err
		}
	}

	return &Provider[A, B]{
		factory: factory,
		decoder: decoder,
		client:  c,
	}, nil
}

// Prepares new Batch inference job
func (p *Provider[A, B]) Prepare() (*Job[A, B], error) {
	g := guid.G(guid.Clock)
	t := guid.EpochT(g)
	u := guid.Base62(g)
	path := fmt.Sprintf("/%04d/%02d/%02d", t.Year(), t.Month(), t.Day())
	file := fmt.Sprintf("%s.jsonl", u)

	fd, err := p.client.fs.Create(filepath.Join(path, file), nil)
	if err != nil {
		return nil, err
	}

	input := fmt.Sprintf("s3://%s%s/%s", p.client.fs.bucket, path, file)
	reply := fmt.Sprintf("s3://%s%s/", p.client.fs.bucket, path)

	spec := &bedrock.CreateModelInvocationJobInput{
		RoleArn: aws.String(p.client.fs.role),
		InputDataConfig: &types.ModelInvocationJobInputDataConfigMemberS3InputDataConfig{
			Value: types.ModelInvocationJobS3InputDataConfig{
				S3Uri: aws.String(input),
			},
		},
		OutputDataConfig: &types.ModelInvocationJobOutputDataConfigMemberS3OutputDataConfig{
			Value: types.ModelInvocationJobS3OutputDataConfig{
				S3Uri: aws.String(reply),
			},
		},
		ModelId: aws.String(p.client.model),
		JobName: aws.String(u),
	}

	srv := provider.New(p.factory, decoder[B]{}, &job[A, B]{w: json.NewEncoder(fd)})
	job := &Job[A, B]{
		Provider: srv,
		fd:       fd,
		uid:      u,
		api:      p.client.api,
		job:      spec,
	}

	return job, nil
}

type Job[A, B any] struct {
	*provider.Provider[A, B]
	fd stream.File

	uid string
	api Runtime
	job *bedrock.CreateModelInvocationJobInput
}

func (job *Job[A, B]) Commit() (string, error) {
	if err := job.fd.Close(); err != nil {
		return "", err
	}

	val, err := job.api.CreateModelInvocationJob(context.Background(), job.job)
	if err != nil {
		return "", err
	}

	return aws.ToString(val.JobArn), nil
}

func (job *Job[A, B]) Cancel() error {
	if err := job.fd.Cancel(); err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

// Batch Inference API
// https://docs.aws.amazon.com/bedrock/latest/userguide/batch-inference-example.html
type batchInput struct {
	ID    string          `json:"recordId,omitempty"`
	Input json.RawMessage `json:"modelInput"`
}

// type batchReply struct {
// 	ID     string          `json:"recordId,omitempty"`
// 	Input  json.RawMessage `json:"modelInput"`
// 	Output json.RawMessage `json:"modelOutput"`
// }

type job[A, B any] struct {
	w *json.Encoder
}

func (s *job[A, B]) Invoke(ctx context.Context, input A) (B, error) {
	req, err := json.Marshal(input)
	if err != nil {
		return *new(B), err
	}

	inquiry := batchInput{
		ID:    guid.Base62(guid.G(guid.Clock)),
		Input: req,
	}

	err = s.w.Encode(&inquiry)
	if err != nil {
		return *new(B), err
	}

	return *new(B), nil
}

type decoder[B any] struct{}

func (decoder decoder[B]) Decode(reply B) (*chatter.Reply, error) {
	return &chatter.Reply{Stage: chatter.LLM_RETURN}, nil
}
