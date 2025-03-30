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
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/fogfish/guid/v2"
	"github.com/fogfish/opts"
	"github.com/fogfish/stream"
	"github.com/kshard/chatter"
	bedrockapi "github.com/kshard/chatter/llm/bedrock"
)

type Client struct {
	llm     bedrockapi.LLM
	bucket  string
	role    string
	fsys    FileSystem
	bedrock Bedrock
}

// Create new AWS Bedrock batch inference client
func New(opt ...Option) (*Client, error) {
	c := Client{}

	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}

	if c.bedrock == nil {
		if err := optsFromRegion(&c, "us-west-2"); err != nil {
			return nil, err
		}
	}

	return &c, c.checkRequired()
}

// Prepare the inference job.
func (c *Client) Prepare() (*Job, error) {
	g := guid.G(guid.Clock)
	t := guid.EpochT(g)
	u := guid.Base62(g)
	path := fmt.Sprintf("/%04d/%02d/%02d", t.Year(), t.Month(), t.Day())
	file := fmt.Sprintf("%s.jsonl", u)
	input := fmt.Sprintf("s3://%s%s/%s", c.bucket, path, file)
	reply := fmt.Sprintf("s3://%s%s/", c.bucket, path)

	fd, err := c.fsys.Create(filepath.Join(path, file), nil)
	if err != nil {
		return nil, err
	}

	spec := &bedrock.CreateModelInvocationJobInput{
		RoleArn: aws.String(c.role),
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
		ModelId: aws.String(c.llm.ModelID()),
		JobName: aws.String(u),
	}

	return &Job{
		writer: newWriter(c.llm, fd),

		uid: u,
		api: c.bedrock,
		job: spec,
	}, nil
}

//------------------------------------------------------------------------------

type writer struct {
	llm   bedrockapi.LLM
	fd    stream.File
	codec *json.Encoder
	tUsed int
}

func newWriter(llm bedrockapi.LLM, fd stream.File) *writer {
	return &writer{
		llm:   llm,
		fd:    fd,
		codec: json.NewEncoder(fd),
		tUsed: 0,
	}
}

func (w *writer) UsedInputTokens() int { return w.tUsed }
func (w *writer) UsedReplyTokens() int { return 0 }

func (w *writer) Prompt(
	ctx context.Context,
	prompt []fmt.Stringer,
	opts ...chatter.Opt,
) (chatter.Reply, error) {
	if w.codec == nil {
		return chatter.Reply{}, fmt.Errorf("job is closed, unable to prompt")
	}

	req, err := w.llm.Encode(prompt, opts...)
	if err != nil {
		return chatter.Reply{}, err
	}

	w.tUsed += int(float64(len(req)) / 4.5)

	inquiry := batchInquiry{
		ID:    guid.Base62(guid.G(guid.Clock)),
		Input: req,
	}

	if err := w.codec.Encode(&inquiry); err != nil {
		return chatter.Reply{}, err
	}

	return chatter.Reply{}, nil
}

func (w *writer) Cancel() error {
	if w.codec == nil {
		return fmt.Errorf("job is closed, unable to cancel")
	}
	w.codec = nil

	w.fd.Cancel()
	w.fd = nil

	return nil
}

func (w *writer) Close() error {
	if w.codec == nil {
		return fmt.Errorf("job is closed, unable to close")
	}

	if err := w.fd.Close(); err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------

type Job struct {
	*writer

	uid string
	api Bedrock
	job *bedrock.CreateModelInvocationJobInput
}

var _ chatter.Chatter = (*Job)(nil)

// Batch Inference API
// https://docs.aws.amazon.com/bedrock/latest/userguide/batch-inference-example.html
type batchInquiry struct {
	ID    string          `json:"recordId,omitempty"`
	Input json.RawMessage `json:"modelInput"`
}

func (job *Job) ID() string { return job.uid }

func (job *Job) Commit() (string, error) {
	if err := job.writer.Close(); err != nil {
		return "", err
	}

	val, err := job.api.CreateModelInvocationJob(context.Background(), job.job)
	if err != nil {
		return "", err
	}

	return aws.ToString(val.JobArn), nil
}

func (job *Job) Cancel() error {
	if err := job.writer.Cancel(); err != nil {
		return err
	}

	return nil
}
