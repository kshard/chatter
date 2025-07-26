//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package provider_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

// Mock input type for testing
type mockInput struct {
	inferrer provider.Inferrer
	commands []chatter.Cmd
	messages []string
	error    error
}

// Mock output type for testing
type mockOutput struct {
	content string
	tokens  chatter.Usage
}

// Mock encoder implementation
type mockEncoder struct {
	input *mockInput
}

func (e *mockEncoder) WithInferrer(config provider.Inferrer) {
	e.input.inferrer = config
}

func (e *mockEncoder) WithCommand(cmd chatter.Cmd) {
	e.input.commands = append(e.input.commands, cmd)
}

func (e *mockEncoder) AsStratum(stratum chatter.Stratum) error {
	if e.input.error != nil {
		return e.input.error
	}
	e.input.messages = append(e.input.messages, "stratum:"+string(stratum))
	return nil
}

func (e *mockEncoder) AsText(text chatter.Text) error {
	if e.input.error != nil {
		return e.input.error
	}
	e.input.messages = append(e.input.messages, "text:"+string(text))
	return nil
}

func (e *mockEncoder) AsPrompt(prompt *chatter.Prompt) error {
	if e.input.error != nil {
		return e.input.error
	}
	e.input.messages = append(e.input.messages, "prompt:"+prompt.String())
	return nil
}

func (e *mockEncoder) AsAnswer(answer *chatter.Answer) error {
	if e.input.error != nil {
		return e.input.error
	}
	e.input.messages = append(e.input.messages, "answer:"+answer.String())
	return nil
}

func (e *mockEncoder) AsReply(reply *chatter.Reply) error {
	if e.input.error != nil {
		return e.input.error
	}
	e.input.messages = append(e.input.messages, "reply:"+reply.String())
	return nil
}

func (e *mockEncoder) Build() *mockInput {
	return e.input
}

// Mock factory
type mockFactory struct {
	error error
}

func (f *mockFactory) Create() (provider.Encoder[*mockInput], error) {
	if f.error != nil {
		return nil, f.error
	}
	return &mockEncoder{
		input: &mockInput{
			messages: make([]string, 0),
			commands: make([]chatter.Cmd, 0),
		},
	}, nil
}

// Mock decoder
type mockDecoder struct {
	error error
}

func (d *mockDecoder) Decode(output *mockOutput) (*chatter.Reply, error) {
	if d.error != nil {
		return nil, d.error
	}
	return &chatter.Reply{
		Stage:   chatter.LLM_RETURN,
		Usage:   output.tokens,
		Content: []chatter.Content{chatter.Text(output.content)},
	}, nil
}

// Mock service
type mockService struct {
	error  error
	output *mockOutput
}

func (s *mockService) Invoke(ctx context.Context, input *mockInput) (*mockOutput, error) {
	if s.error != nil {
		return nil, s.error
	}
	return s.output, nil
}

func TestProvider_PromptEmptyError(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoder := &mockDecoder{}
	service := &mockService{
		output: &mockOutput{
			content: "test response",
			tokens:  chatter.Usage{InputTokens: 10, ReplyTokens: 20},
		},
	}

	p := provider.New(factory, decoder, service)

	it.Then(t).Should(
		it.Error(p.Prompt(context.Background(), []chatter.Message{})).Contain("empty prompt"),
	)
}

func TestProvider_PromptFactoryError(t *testing.T) {
	factoryErr := errors.New("factory error")
	factory := func() (provider.Encoder[*mockInput], error) {
		return nil, factoryErr
	}
	decoder := &mockDecoder{}
	service := &mockService{}

	p := provider.New(factory, decoder, service)

	it.Then(t).Should(
		it.Error(p.Prompt(context.Background(), []chatter.Message{chatter.Text("test")})).Contain("factory error"),
	)
}

func TestProvider_PromptServiceError(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoder := &mockDecoder{}
	serviceErr := errors.New("service error")
	service := &mockService{error: serviceErr}

	p := provider.New(factory, decoder, service)

	it.Then(t).Should(
		it.Error(p.Prompt(context.Background(), []chatter.Message{chatter.Text("test")})).Contain("service error"),
	)
}

func TestProvider_PromptDecoderError(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoderErr := errors.New("decoder error")
	decoder := &mockDecoder{error: decoderErr}
	service := &mockService{
		output: &mockOutput{content: "test"},
	}

	p := provider.New(factory, decoder, service)

	it.Then(t).Should(
		it.Error(p.Prompt(context.Background(), []chatter.Message{chatter.Text("test")})).Contain("decoder error"),
	)
}

func TestProvider_PromptEncoderError(t *testing.T) {
	encoderErr := errors.New("encoder error")
	factory := func() (provider.Encoder[*mockInput], error) {
		return &mockEncoder{
			input: &mockInput{error: encoderErr},
		}, nil
	}
	decoder := &mockDecoder{}
	service := &mockService{}

	p := provider.New(factory, decoder, service)

	it.Then(t).Should(
		it.Error(p.Prompt(context.Background(), []chatter.Message{chatter.Text("test")})).Contain("encoder error"),
	)
}

func TestProvider_PromptMessageTypes(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoder := &mockDecoder{}
	service := &mockService{
		output: &mockOutput{
			content: "response",
			tokens:  chatter.Usage{InputTokens: 15, ReplyTokens: 25},
		},
	}

	p := provider.New(factory, decoder, service)

	// Create test messages
	prompt := &chatter.Prompt{}
	prompt.WithTask("Test task")
	answer := &chatter.Answer{Yield: []chatter.Json{}}
	reply := &chatter.Reply{Content: []chatter.Content{chatter.Text("previous")}}

	messages := []chatter.Message{
		chatter.Stratum("system role"),
		chatter.Text("user text"),
		prompt,
		answer,
		reply,
	}

	result, err := p.Prompt(context.Background(), messages)

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(result.Stage, chatter.LLM_RETURN),
		it.Equal(result.Usage.InputTokens, 15),
		it.Equal(result.Usage.ReplyTokens, 25),
		it.Equal(result.String(), "response"),
	)

	// Verify usage tracking
	it.Then(t).Should(
		it.Equal(p.Usage().InputTokens, 15),
		it.Equal(p.Usage().ReplyTokens, 25),
	)
}

func TestProvider_PromptWithOptions(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoder := &mockDecoder{}
	service := &mockService{
		output: &mockOutput{
			content: "response",
			tokens:  chatter.Usage{InputTokens: 10, ReplyTokens: 20},
		},
	}

	p := provider.New(factory, decoder, service)

	opts := []chatter.Opt{
		chatter.Temperature(0.7),
		chatter.TopP(0.9),
		chatter.TopK(50),
		chatter.MaxTokens(1000),
		chatter.StopSequences([]string{"stop1", "stop2"}),
		chatter.Registry([]chatter.Cmd{
			{Cmd: "test_cmd", About: "Test command", Schema: []byte(`{"type":"object"}`)},
		}),
	}

	reply, err := p.Prompt(context.Background(), []chatter.Message{chatter.Text("test")}, opts...)

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(reply.String(), "response"),
	)
}

func TestProvider_PromptBasicFlow(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoder := &mockDecoder{}
	service := &mockService{
		output: &mockOutput{
			content: "Hello, World!",
			tokens:  chatter.Usage{InputTokens: 5, ReplyTokens: 10},
		},
	}

	p := provider.New(factory, decoder, service)

	reply, err := p.Prompt(context.Background(), []chatter.Message{chatter.Text("Hello")})

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(reply.Stage, chatter.LLM_RETURN),
		it.Equal(reply.String(), "Hello, World!"),
		it.Equal(reply.Usage.InputTokens, 5),
		it.Equal(reply.Usage.ReplyTokens, 10),
	)

	// Test cumulative usage tracking
	_, err = p.Prompt(context.Background(), []chatter.Message{chatter.Text("Hello again")})

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(p.Usage().InputTokens, 10), // 5 + 5
		it.Equal(p.Usage().ReplyTokens, 20), // 10 + 10
	)
}

func TestProvider_PromptWithComplexPrompt(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoder := &mockDecoder{}
	service := &mockService{
		output: &mockOutput{
			content: "Complex response",
			tokens:  chatter.Usage{InputTokens: 100, ReplyTokens: 150},
		},
	}

	p := provider.New(factory, decoder, service)

	// Create a complex prompt
	prompt := &chatter.Prompt{}
	prompt.WithTask("Analyze the given text")
	prompt.WithRules("Rule 1", "Be concise", "Be accurate")
	prompt.WithContext("Context", "This is important data")
	prompt.WithInput("Input", "Sample text to analyze")

	messages := []chatter.Message{
		chatter.Stratum("You are a helpful assistant"),
		prompt,
	}

	opts := []chatter.Opt{
		chatter.Temperature(0.5),
		chatter.MaxTokens(500),
	}

	reply, err := p.Prompt(context.Background(), messages, opts...)

	it.Then(t).Should(
		it.Nil(err),
		it.Equal(reply.Stage, chatter.LLM_RETURN),
		it.Equal(reply.String(), "Complex response"),
		it.Equal(reply.Usage.InputTokens, 100),
		it.Equal(reply.Usage.ReplyTokens, 150),
	)
}

func TestProvider_Usage(t *testing.T) {
	factory := func() (provider.Encoder[*mockInput], error) {
		return (&mockFactory{}).Create()
	}
	decoder := &mockDecoder{}
	service := &mockService{
		output: &mockOutput{
			content: "test",
			tokens:  chatter.Usage{InputTokens: 42, ReplyTokens: 84},
		},
	}

	p := provider.New(factory, decoder, service)

	// Initially, usage should be zero
	usage := p.Usage()
	it.Then(t).Should(
		it.Equal(usage.InputTokens, 0),
		it.Equal(usage.ReplyTokens, 0),
	)

	// After a prompt, usage should be updated
	_, err := p.Prompt(context.Background(), []chatter.Message{chatter.Text("test")})
	it.Then(t).Should(it.Nil(err))

	usage = p.Usage()
	it.Then(t).Should(
		it.Equal(usage.InputTokens, 42),
		it.Equal(usage.ReplyTokens, 84),
	)
}
