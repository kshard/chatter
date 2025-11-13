//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package imagen

import (
	"context"
	"strings"

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

type input struct {
	Model  string
	Prompt strings.Builder
}

type Imagen = provider.Provider[*input, *genai.GenerateImagesResponse]

func New(model string, opt Config) (*Imagen, error) {
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

var _ provider.Service[*input, *genai.GenerateImagesResponse] = (*Service)(nil)

func (s *Service) Invoke(ctx context.Context, input *input) (*genai.GenerateImagesResponse, error) {
	return s.api.Models.GenerateImages(ctx, input.Model, input.Prompt.String(), &genai.GenerateImagesConfig{})
}
