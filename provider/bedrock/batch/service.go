//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package batch

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"

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

// AWS Bedrock control-plane API surface required to submit and track
// a batch inference job.
type Runtime interface {
	CreateModelInvocationJob(ctx context.Context, params *bedrock.CreateModelInvocationJobInput, optFns ...func(*bedrock.Options)) (*bedrock.CreateModelInvocationJobOutput, error)
	GetModelInvocationJob(ctx context.Context, params *bedrock.GetModelInvocationJobInput, optFns ...func(*bedrock.Options)) (*bedrock.GetModelInvocationJobOutput, error)
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

// Prepares new Batch inference job. The returned Job accepts prompts via
// [Job.Prompt], which enqueues them into the job's input file and returns
// a record identifier. Once [Job.Commit] submits the job, the same
// identifiers are used by [Job.Fetch] to correlate decoded results with
// the request that produced them.
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

	svc := &writer[A, B]{w: json.NewEncoder(fd)}
	inner := provider.New(p.factory, noop[B]{}, svc)

	job := &Job[A, B]{
		Provider: inner,
		svc:      svc,
		fd:       fd,
		fs:       p.client.fs,
		path:     path,
		file:     file,
		uid:      u,
		api:      p.client.api,
		decoder:  p.decoder,
		spec:     spec,
	}

	return job, nil
}

// Lifecycle state of a [Job]. A job accepts prompts only while open,
// and can be committed or canceled exactly once.
type jobState int

const (
	jobOpen jobState = iota
	jobCommitted
	jobCanceled
)

func (s jobState) String() string {
	switch s {
	case jobCommitted:
		return "committed"
	case jobCanceled:
		return "canceled"
	default:
		return "open"
	}
}

// Job represents a single batch inference job: a JSONL file of requests
// written to S3, submitted to Bedrock as a single CreateModelInvocationJob.
type Job[A, B any] struct {
	*provider.Provider[A, B]

	mu    sync.Mutex
	state jobState
	svc   *writer[A, B]

	fd   stream.File
	fs   fs.FS
	path string
	file string

	uid     string
	arn     string
	api     Runtime
	decoder provider.Decoder[B]
	spec    *bedrock.CreateModelInvocationJobInput
}

// Prompt encodes the sequence of messages into a single batch record and
// appends it to the job's input file. Unlike [chatter.Chatter], it does not
// return a completion: batch inference is asynchronous, so it returns the
// record identifier instead. Use that identifier to correlate this request
// with the corresponding [Result] once [Job.Fetch] reads back the job's
// output, after [Job.Commit] and completion of the job.
func (job *Job[A, B]) Prompt(ctx context.Context, prompt []chatter.Message, opts ...chatter.Opt) (string, error) {
	job.mu.Lock()
	defer job.mu.Unlock()

	if job.state != jobOpen {
		return "", fmt.Errorf("job is %s, no more records can be enqueued", job.state)
	}

	if _, err := job.Provider.Prompt(ctx, prompt, opts...); err != nil {
		return "", err
	}

	return job.svc.id, nil
}

// Commit closes the job's input file and submits it to Bedrock for
// batch inference. It returns the ARN of the submitted job, later used
// by [Job.Status] and [Job.Fetch].
func (job *Job[A, B]) Commit() (string, error) {
	job.mu.Lock()
	defer job.mu.Unlock()

	if job.state != jobOpen {
		return "", fmt.Errorf("job is %s, it cannot be committed", job.state)
	}

	if err := job.fd.Close(); err != nil {
		return "", err
	}

	val, err := job.api.CreateModelInvocationJob(context.Background(), job.spec)
	if err != nil {
		return "", err
	}

	job.state = jobCommitted
	job.arn = aws.ToString(val.JobArn)

	return job.arn, nil
}

// Cancel aborts the job's input file upload. It is only valid before
// the job has been committed.
func (job *Job[A, B]) Cancel() error {
	job.mu.Lock()
	defer job.mu.Unlock()

	if job.state != jobOpen {
		return fmt.Errorf("job is %s, it cannot be canceled", job.state)
	}

	if err := job.fd.Cancel(); err != nil {
		return err
	}

	job.state = jobCanceled

	return nil
}

// Status queries Bedrock for the current status of a committed job
// (e.g. InProgress, Completed, Failed). See [types.ModelInvocationJobStatus].
func (job *Job[A, B]) Status(ctx context.Context) (types.ModelInvocationJobStatus, error) {
	job.mu.Lock()
	arn, state := job.arn, job.state
	job.mu.Unlock()

	if state != jobCommitted {
		return "", fmt.Errorf("job is %s, it has not been submitted yet", state)
	}

	val, err := job.api.GetModelInvocationJob(ctx, &bedrock.GetModelInvocationJobInput{
		JobIdentifier: aws.String(arn),
	})
	if err != nil {
		return "", err
	}

	return val.Status, nil
}

// Result is the decoded outcome of a single batch record. ID matches the
// identifier returned earlier from [Job.Prompt], allowing the caller to
// correlate results with the requests that produced them.
type Result struct {
	ID    string
	Reply *chatter.Reply
	Err   error
}

// Fetch downloads and decodes the output of a job. It should be called
// once the job has reached a terminal status (Completed or
// PartiallyCompleted, see [Job.Status]); records that Bedrock could not
// process are returned with Err set instead of Reply.
func (job *Job[A, B]) Fetch(ctx context.Context) ([]Result, error) {
	job.mu.Lock()
	arn, state := job.arn, job.state
	job.mu.Unlock()

	if state != jobCommitted {
		return nil, fmt.Errorf("job is %s, it has not been submitted yet", state)
	}

	id := arn
	if n := strings.LastIndex(id, "/"); n >= 0 {
		id = id[n+1:]
	}

	fd, err := job.fs.Open(filepath.Join(job.path, id, job.file+".out"))
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	seq := make([]Result, 0)

	scanner := bufio.NewScanner(fd)
	scanner.Buffer(make([]byte, 0, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		var rec batchOutput
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			return nil, err
		}

		if rec.Error != nil {
			seq = append(seq, Result{ID: rec.ID, Err: fmt.Errorf("%s (code %d)", rec.Error.Message, rec.Error.Code)})
			continue
		}

		var native B
		if err := json.Unmarshal(rec.Output, &native); err != nil {
			return nil, err
		}

		reply, err := job.decoder.Decode(native)
		if err != nil {
			return nil, err
		}

		seq = append(seq, Result{ID: rec.ID, Reply: reply})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return seq, nil
}

//------------------------------------------------------------------------------

// Batch Inference API
// https://docs.aws.amazon.com/bedrock/latest/userguide/batch-inference-example.html
type batchInput struct {
	ID    string          `json:"recordId,omitempty"`
	Input json.RawMessage `json:"modelInput"`
}

type batchOutput struct {
	ID     string            `json:"recordId,omitempty"`
	Output json.RawMessage   `json:"modelOutput,omitempty"`
	Error  *batchOutputError `json:"error,omitempty"`
}

type batchOutputError struct {
	Code    int    `json:"errorCode"`
	Message string `json:"errorMessage"`
}

// writer implements provider.Service[A, B]. It serializes each encoded
// request into the job's JSONL file, tagged with a freshly minted record
// identifier that [Job.Prompt] hands back to the caller for correlation.
type writer[A, B any] struct {
	w  *json.Encoder
	id string
}

func (s *writer[A, B]) Invoke(ctx context.Context, input A) (B, error) {
	req, err := json.Marshal(input)
	if err != nil {
		return *new(B), err
	}

	id := guid.Base62(guid.G(guid.Clock))
	if err := s.w.Encode(&batchInput{ID: id, Input: req}); err != nil {
		return *new(B), err
	}

	s.id = id

	return *new(B), nil
}

// noop implements provider.Decoder[B] for the write phase, where no real
// completion exists yet. Real decoding of Bedrock's output happens in
// [Job.Fetch], using the decoder supplied to [New].
type noop[B any] struct{}

func (noop[B]) Decode(B) (*chatter.Reply, error) {
	return &chatter.Reply{Stage: chatter.LLM_RETURN}, nil
}
