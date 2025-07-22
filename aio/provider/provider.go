package provider

import (
	"context"
	"fmt"

	"github.com/fogfish/faults"
	"github.com/kshard/chatter"
)

const (
	ErrBadRequest = faults.Type("bad request")
	ErrServiceIO  = faults.Type("service IO error")
)

type Factory[A any] func() (Encoder[A], error)

type Inferrer interface {
	WithTemperature(float64)
	WithTopP(float64)
	WithMaxTokens(int)
	WithStopSequences([]string)
	WithCommand(chatter.Cmd)
}

type Prompter interface {
	AsStratum(chatter.Stratum) error
	AsText(chatter.Text) error
	AsPrompt(*chatter.Prompt) error
	AsAnswer(*chatter.Answer) error
	AsReply(*chatter.Reply) error
}

type Encoder[A any] interface {
	Inferrer
	Prompter
	Build() A
}

type Decoder[B any] interface {
	Decode(B) (*chatter.Reply, error)
}

type Service[A, B any] interface {
	Invoke(context.Context, A) (B, error)
}

type Provider[A, B any] struct {
	factory Factory[A]
	decoder Decoder[B]
	service Service[A, B]

	usage chatter.Usage
}

func New[A, B any](
	factory Factory[A],
	decoder Decoder[B],
	service Service[A, B],
) *Provider[A, B] {
	return &Provider[A, B]{
		factory: factory,
		decoder: decoder,
		service: service,
	}
}

func (p *Provider[A, B]) Usage() chatter.Usage { return p.usage }

func (p *Provider[A, B]) Prompt(ctx context.Context, prompt []chatter.Message, opts ...chatter.Opt) (*chatter.Reply, error) {
	if len(prompt) == 0 {
		return nil, ErrBadRequest.With(fmt.Errorf("empty prompt"))
	}

	input, err := p.factory()
	if err != nil {
		return nil, ErrBadRequest.With(err)
	}

	for _, opt := range opts {
		switch v := opt.(type) {
		case chatter.Temperature:
			input.WithTemperature(float64(v))
		case chatter.TopP:
			input.WithTopP(float64(v))
		case chatter.Quota:
			input.WithMaxTokens(int(v))
		case chatter.StopSequences:
			input.WithStopSequences([]string(v))
		case chatter.Registry:
			for _, cmd := range v {
				input.WithCommand(cmd)
			}
		}
	}

	for _, term := range prompt {
		switch v := term.(type) {
		case chatter.Stratum:
			if err := input.AsStratum(v); err != nil {
				return nil, ErrBadRequest.With(err)
			}
		case chatter.Text:
			if err := input.AsText(v); err != nil {
				return nil, ErrBadRequest.With(err)
			}
		case *chatter.Prompt:
			if err := input.AsPrompt(v); err != nil {
				return nil, ErrBadRequest.With(err)
			}
		case *chatter.Answer:
			if err := input.AsAnswer(v); err != nil {
				return nil, ErrBadRequest.With(err)
			}
		case *chatter.Reply:
			if err := input.AsReply(v); err != nil {
				return nil, ErrBadRequest.With(err)
			}
		default:
			return nil, ErrBadRequest.With(fmt.Errorf("unsupported message type %T", term))
		}
	}

	req := input.Build()
	result, err := p.service.Invoke(ctx, req)
	if err != nil {
		return nil, ErrServiceIO.With(err)
	}

	reply, err := p.decoder.Decode(result)
	if err != nil {
		return nil, ErrServiceIO.With(err)
	}

	p.usage.InputTokens += reply.Usage.InputTokens
	p.usage.ReplyTokens += reply.Usage.ReplyTokens

	return reply, nil
}
