//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package batch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/fogfish/it/v2"
	"github.com/fogfish/opts"
	"github.com/fogfish/stream"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

//------------------------------------------------------------------------------
// test doubles
//------------------------------------------------------------------------------

type mockInput struct {
	Text string `json:"text"`
}

type mockOutput struct {
	Text string `json:"text"`
}

type mockEncoder struct{ req *mockInput }

func (e *mockEncoder) WithInferrer(provider.Inferrer)  {}
func (e *mockEncoder) WithCommand(chatter.Cmd)         {}
func (e *mockEncoder) AsStratum(chatter.Stratum) error { return nil }
func (e *mockEncoder) AsText(t chatter.Text) error     { e.req.Text = string(t); return nil }
func (e *mockEncoder) AsPrompt(*chatter.Prompt) error  { return nil }
func (e *mockEncoder) AsAnswer(*chatter.Answer) error  { return nil }
func (e *mockEncoder) AsReply(*chatter.Reply) error    { return nil }
func (e *mockEncoder) Build() *mockInput               { return e.req }

func mockFactory() (provider.Encoder[*mockInput], error) {
	return &mockEncoder{req: &mockInput{}}, nil
}

type mockDecoder struct{}

func (mockDecoder) Decode(out *mockOutput) (*chatter.Reply, error) {
	return &chatter.Reply{
		Stage:   chatter.LLM_RETURN,
		Content: []chatter.Content{chatter.Text(out.Text)},
	}, nil
}

// fakeFile implements stream.File without touching S3.
type fakeFile struct {
	buf      bytes.Buffer
	closed   bool
	canceled bool
}

func (f *fakeFile) Write(p []byte) (int, error) { return f.buf.Write(p) }
func (f *fakeFile) Close() error                { f.closed = true; return nil }
func (f *fakeFile) Cancel() error               { f.canceled = true; return nil }
func (f *fakeFile) Stat() (fs.FileInfo, error)  { return nil, nil }

// fakeRuntime implements Runtime without calling AWS.
type fakeRuntime struct {
	createCalls int
	createFn    func(*bedrock.CreateModelInvocationJobInput) (*bedrock.CreateModelInvocationJobOutput, error)
	statusFn    func(*bedrock.GetModelInvocationJobInput) (*bedrock.GetModelInvocationJobOutput, error)
}

func (r *fakeRuntime) CreateModelInvocationJob(ctx context.Context, in *bedrock.CreateModelInvocationJobInput, optFns ...func(*bedrock.Options)) (*bedrock.CreateModelInvocationJobOutput, error) {
	r.createCalls++
	return r.createFn(in)
}

func (r *fakeRuntime) GetModelInvocationJob(ctx context.Context, in *bedrock.GetModelInvocationJobInput, optFns ...func(*bedrock.Options)) (*bedrock.GetModelInvocationJobOutput, error) {
	return r.statusFn(in)
}

// memFS implements fs.FS backed by an in-memory map, for testing Fetch
// without reading from S3.
type memFS map[string]string

func (m memFS) Open(name string) (fs.File, error) {
	data, ok := m[name]
	if !ok {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return &memFile{Reader: bytes.NewReader([]byte(data))}, nil
}

type memFile struct{ *bytes.Reader }

func (f *memFile) Stat() (fs.FileInfo, error) { return nil, nil }
func (f *memFile) Close() error               { return nil }

// memCreateFS implements stream.CreateFS[struct{}] so Prepare can be
// exercised without a real S3-backed FileSystem.
type memCreateFS struct{}

func (memCreateFS) Open(name string) (fs.File, error) { return nil, fs.ErrNotExist }
func (memCreateFS) Create(path string, attr *struct{}) (stream.File, error) {
	return &fakeFile{}, nil
}

//------------------------------------------------------------------------------
// helpers
//------------------------------------------------------------------------------

func newTestJob() (*Job[*mockInput, *mockOutput], *fakeFile) {
	fd := &fakeFile{}
	svc := &writer[*mockInput, *mockOutput]{w: json.NewEncoder(fd)}
	inner := provider.New(mockFactory, noop[*mockOutput]{}, svc)

	job := &Job[*mockInput, *mockOutput]{
		Provider: inner,
		svc:      svc,
		fd:       fd,
		path:     "/2026/08/25",
		file:     "input.jsonl",
		decoder:  mockDecoder{},
		spec:     &bedrock.CreateModelInvocationJobInput{JobName: aws.String("test")},
	}

	return job, fd
}

//------------------------------------------------------------------------------
// Prompt / record correlation
//------------------------------------------------------------------------------

func TestJob_PromptReturnsRecordID(t *testing.T) {
	job, fd := newTestJob()

	id, err := job.Prompt(context.Background(), []chatter.Message{chatter.Text("hello")})
	it.Then(t).Should(
		it.Nil(err),
		it.True(id != ""),
	)

	var rec batchInput
	it.Then(t).Should(
		it.Nil(json.Unmarshal(fd.buf.Bytes(), &rec)),
		it.Equal(rec.ID, id),
	)

	var in mockInput
	it.Then(t).Should(
		it.Nil(json.Unmarshal(rec.Input, &in)),
		it.Equal(in.Text, "hello"),
	)
}

func TestJob_PromptReturnsUniqueRecordIDs(t *testing.T) {
	job, _ := newTestJob()

	id1, err1 := job.Prompt(context.Background(), []chatter.Message{chatter.Text("a")})
	id2, err2 := job.Prompt(context.Background(), []chatter.Message{chatter.Text("b")})

	it.Then(t).Should(
		it.Nil(err1),
		it.Nil(err2),
		it.True(id1 != id2),
	)
}

func TestJob_PromptAfterCommitFails(t *testing.T) {
	job, _ := newTestJob()
	job.state = jobCommitted

	_, err := job.Prompt(context.Background(), []chatter.Message{chatter.Text("hello")})
	it.Then(t).Should(
		it.Error(struct{}{}, err).Contain("committed"),
	)
}

func TestJob_PromptAfterCancelFails(t *testing.T) {
	job, _ := newTestJob()
	job.state = jobCanceled

	_, err := job.Prompt(context.Background(), []chatter.Message{chatter.Text("hello")})
	it.Then(t).Should(
		it.Error(struct{}{}, err).Contain("canceled"),
	)
}

//------------------------------------------------------------------------------
// Commit / Cancel state machine
//------------------------------------------------------------------------------

func TestJob_CommitClosesFileAndSubmitsJob(t *testing.T) {
	job, fd := newTestJob()
	job.api = &fakeRuntime{
		createFn: func(in *bedrock.CreateModelInvocationJobInput) (*bedrock.CreateModelInvocationJobOutput, error) {
			return &bedrock.CreateModelInvocationJobOutput{
				JobArn: aws.String("arn:aws:bedrock:us-west-2:000000000000:model-invocation-job/abc123"),
			}, nil
		},
	}

	arn, err := job.Commit()
	it.Then(t).Should(
		it.Nil(err),
		it.Equal(arn, "arn:aws:bedrock:us-west-2:000000000000:model-invocation-job/abc123"),
		it.True(fd.closed),
		it.Equal(job.state, jobCommitted),
	)
}

func TestJob_CommitTwiceFails(t *testing.T) {
	job, _ := newTestJob()
	job.api = &fakeRuntime{
		createFn: func(in *bedrock.CreateModelInvocationJobInput) (*bedrock.CreateModelInvocationJobOutput, error) {
			return &bedrock.CreateModelInvocationJobOutput{JobArn: aws.String("arn:job/1")}, nil
		},
	}

	_, err := job.Commit()
	it.Then(t).Should(it.Nil(err))

	_, err = job.Commit()
	it.Then(t).Should(
		it.Error(struct{}{}, err).Contain("committed"),
	)
}

func TestJob_CommitServiceErrorKeepsJobOpen(t *testing.T) {
	job, _ := newTestJob()
	svcErr := errors.New("service unavailable")
	job.api = &fakeRuntime{
		createFn: func(in *bedrock.CreateModelInvocationJobInput) (*bedrock.CreateModelInvocationJobOutput, error) {
			return nil, svcErr
		},
	}

	_, err := job.Commit()
	it.Then(t).Should(
		it.Error(struct{}{}, err).Contain("service unavailable"),
		it.Equal(job.state, jobOpen),
	)
}

func TestJob_CancelBeforeCommit(t *testing.T) {
	job, fd := newTestJob()

	err := job.Cancel()
	it.Then(t).Should(
		it.Nil(err),
		it.True(fd.canceled),
		it.Equal(job.state, jobCanceled),
	)
}

func TestJob_CancelAfterCommitFails(t *testing.T) {
	job, _ := newTestJob()
	job.state = jobCommitted

	err := job.Cancel()
	it.Then(t).Should(
		it.Error(struct{}{}, err).Contain("committed"),
	)
}

//------------------------------------------------------------------------------
// Status
//------------------------------------------------------------------------------

func TestJob_StatusRequiresCommittedJob(t *testing.T) {
	job, _ := newTestJob()

	_, err := job.Status(context.Background())
	it.Then(t).Should(
		it.Error(struct{}{}, err).Contain("open"),
	)
}

func TestJob_StatusReturnsBedrockStatus(t *testing.T) {
	job, _ := newTestJob()
	job.state = jobCommitted
	job.arn = "arn:aws:bedrock:us-west-2:000000000000:model-invocation-job/abc123"
	job.api = &fakeRuntime{
		statusFn: func(in *bedrock.GetModelInvocationJobInput) (*bedrock.GetModelInvocationJobOutput, error) {
			it.Then(t).Should(it.Equal(aws.ToString(in.JobIdentifier), job.arn))
			return &bedrock.GetModelInvocationJobOutput{Status: types.ModelInvocationJobStatusCompleted}, nil
		},
	}

	status, err := job.Status(context.Background())
	it.Then(t).Should(
		it.Nil(err),
		it.Equal(status, types.ModelInvocationJobStatusCompleted),
	)
}

//------------------------------------------------------------------------------
// Fetch
//------------------------------------------------------------------------------

func TestJob_FetchRequiresCommittedJob(t *testing.T) {
	job, _ := newTestJob()

	_, err := job.Fetch(context.Background())
	it.Then(t).Should(
		it.Error(struct{}{}, err).Contain("open"),
	)
}

func TestJob_FetchDecodesAndCorrelatesResults(t *testing.T) {
	job, _ := newTestJob()
	job.state = jobCommitted
	job.arn = "arn:aws:bedrock:us-west-2:000000000000:model-invocation-job/abc123"

	output := strings.Join([]string{
		`{"recordId":"rec-1","modelInput":{"text":"hi"},"modelOutput":{"text":"hello back"}}`,
		`{"recordId":"rec-2","modelInput":{"text":"bad"},"error":{"errorCode":400,"errorMessage":"bad request"}}`,
	}, "\n")

	job.fs = memFS{
		"/2026/08/25/abc123/input.jsonl.out": output,
	}

	results, err := job.Fetch(context.Background())
	it.Then(t).Should(
		it.Nil(err),
		it.Equal(len(results), 2),
	)

	it.Then(t).Should(
		it.Equal(results[0].ID, "rec-1"),
		it.Nil(results[0].Err),
		it.Equal(results[0].Reply.String(), "hello back"),
	)

	it.Then(t).Should(
		it.Equal(results[1].ID, "rec-2"),
		it.True(results[1].Reply == nil),
		it.Error(struct{}{}, results[1].Err).Contain("bad request"),
	)
}

func TestJob_FetchMissingOutputFails(t *testing.T) {
	job, _ := newTestJob()
	job.state = jobCommitted
	job.arn = "arn:aws:bedrock:us-west-2:000000000000:model-invocation-job/abc123"
	job.fs = memFS{}

	_, err := job.Fetch(context.Background())
	it.Then(t).Should(it.True(err != nil))
}

//------------------------------------------------------------------------------
// writer
//------------------------------------------------------------------------------

func TestWriter_InvokeEncodesRecordAndKeepsID(t *testing.T) {
	fd := &fakeFile{}
	w := &writer[*mockInput, *mockOutput]{w: json.NewEncoder(fd)}

	_, err := w.Invoke(context.Background(), &mockInput{Text: "hi"})
	it.Then(t).Should(it.Nil(err))

	var rec batchInput
	it.Then(t).Should(it.Nil(json.Unmarshal(fd.buf.Bytes(), &rec)))
	it.Then(t).Should(
		it.True(rec.ID != ""),
		it.Equal(rec.ID, w.id),
	)
}

//------------------------------------------------------------------------------
// options
//------------------------------------------------------------------------------

func TestWithModelInvocationTypeSetsClientField(t *testing.T) {
	c := Client{}
	err := opts.Apply(&c, []Option{WithModelInvocationType(types.ModelInvocationTypeConverse)})

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(c.invocationType, types.ModelInvocationTypeConverse),
	)
}

func TestPrepareSetsModelInvocationTypeOnJobSpec(t *testing.T) {
	p := &Provider[*mockInput, *mockOutput]{
		factory: mockFactory,
		decoder: mockDecoder{},
		client: Client{
			fs:             &FileSystem{CreateFS: memCreateFS{}, bucket: "test-bucket"},
			model:          "test-model",
			invocationType: types.ModelInvocationTypeConverse,
		},
	}

	job, err := p.Prepare()
	it.Then(t).Should(
		it.Nil(err),
		it.Equal(job.spec.ModelInvocationType, types.ModelInvocationTypeConverse),
	)
}
