//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package gemini

import (
	"context"

	"github.com/kshard/chatter/aio/provider"
	"google.golang.org/genai"
)

type Config struct {
	// API secret key
	Secret string
}

type Service struct {
	api *genai.Client
}

type Gemini = provider.Provider[*input, *genai.GenerateContentResponse]

func New(model string, opt Config) (*Gemini, error) {
	config := &genai.ClientConfig{
		APIKey: opt.Secret,
	}

	api, err := genai.NewClient(context.Background(), config)
	if err != nil {
		return nil, err
	}

	c := &Service{api: api}

	return provider.New(factory(model), decoder{}, c), nil
}

//------------------------------------------------------------------------------

var _ provider.Service[*input, *genai.GenerateContentResponse] = (*Service)(nil)

func (s *Service) Invoke(ctx context.Context, input *input) (*genai.GenerateContentResponse, error) {
	return s.api.Models.GenerateContent(ctx, input.Model, input.Prompt, &genai.GenerateContentConfig{})
}
