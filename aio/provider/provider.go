//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package provider

import (
	"context"
	"fmt"

	"github.com/fogfish/faults"
	"github.com/kshard/chatter"
)

//
// Provider is a generic interface for LLMs providers.
// It defines a basic contract for LLMs adapters, allowing to focus on
// implementation of LLMs protocols and their encoder/decoder.
//

const (
	ErrBadRequest = faults.Type("bad request")
	ErrServiceIO  = faults.Type("service IO error")
)

// Encoder Factory is a function that creates an instance of LLM request.
type Factory[A any] func() (Encoder[A], error)

// Inferrer is a set of parameters that influence the LLM's inference behavior.
type Inferrer struct {
	Temperature   float64
	TopP          float64
	TopK          float64
	MaxTokens     int
	StopSequences []string
}

// LLM request encoder.
type Encoder[A any] interface {
	WithInferrer(Inferrer)
	WithCommand(chatter.Cmd)

	AsStratum(chatter.Stratum) error
	AsText(chatter.Text) error
	AsPrompt(*chatter.Prompt) error
	AsAnswer(*chatter.Answer) error
	AsReply(*chatter.Reply) error

	Build() A
}

// LLM response decoder.
type Decoder[B any] interface {
	Decode(B) (*chatter.Reply, error)
}

// Service is a generic I/O for LLMs provider.
type Service[A, B any] interface {
	Invoke(context.Context, A) (B, error)
}

// Provider is a generic implementation of Chatter interface.
type Provider[A, B any] struct {
	factory Factory[A]
	decoder Decoder[B]
	service Service[A, B]

	usage chatter.Usage
}

var _ chatter.Chatter = (*Provider[any, any])(nil)

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

	if len(opts) > 0 {
		var config Inferrer

		for _, opt := range opts {
			switch v := opt.(type) {
			case chatter.Temperature:
				config.Temperature = float64(v)
			case chatter.TopP:
				config.TopP = float64(v)
			case chatter.TopK:
				config.TopK = float64(v)
			case chatter.MaxTokens:
				config.MaxTokens = int(v)
			case chatter.StopSequences:
				config.StopSequences = make([]string, len(v))
				copy(config.StopSequences, v)
			case chatter.Registry:
				for _, cmd := range v {
					input.WithCommand(cmd)
				}
			}
		}
		input.WithInferrer(config)
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
