//
// Copyright (C) 2024 - 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package autoconfig

import (
	"fmt"
	gohttp "net/http"
	"strings"

	"time"

	"github.com/fogfish/gurl/v2/http"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/provider/bedrock"
	"github.com/kshard/chatter/provider/bedrock/embedding/titan"
	"github.com/kshard/chatter/provider/bedrock/foundation/converse"
	"github.com/kshard/chatter/provider/bedrock/foundation/llama"
	"github.com/kshard/chatter/provider/bedrock/foundation/nova"
	"github.com/kshard/chatter/provider/google/foundation/gemini"
	"github.com/kshard/chatter/provider/google/foundation/imagen"
	"github.com/kshard/chatter/provider/openai"
	"github.com/kshard/chatter/provider/openai/embedding/text2vec"
	"github.com/kshard/chatter/provider/openai/foundation/gpt"
)

// Instance of LLM provider configuration, used for automatic configuration of LLM instances.
type Instance struct {
	// Unique name for the instance, used for reference in the application.
	Name string `json:"name" yaml:"name"`

	// Provider's identity and the capability to be used for inference job.
	// The identity path to github.com/kshard/chatter/provider submobile.
	// It consists of three segments: provider, capability, and family.
	//
	// For example, `provider:bedrock/foundation/converse` specifies the provider
	// as Bedrock and the capability as Converse.
	Provider string `json:"provider" yaml:"provider"`

	// Unique model identifier as defined by the provider.
	// For example, `gemini-1.5-pro` for Google Gemini Pro.
	Model string `json:"model" yaml:"model"`

	// Bedrock specific, AWS Region where the model is hosted. For example, `us-east-1`.
	Region string `json:"region,omitempty" yaml:"region,omitempty"`

	// OpenAI specific, API host URL. For example, `https://api.openai.com`.
	Host string `json:"host,omitempty" yaml:"host,omitempty"`

	// OpenAI specific, API secret key. For example, `sk-xxxxxx`.
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty"`

	// Timeout in seconds for API requests. For example, `30`.
	Timeout int `json:"timeout,omitempty" yaml:"timeout,omitempty"`

	// Dimensions for embedding models. For example, `1024` for Titan embedding.
	Dimensions int `json:"dimensions,omitempty" yaml:"dimensions,omitempty"`
}

// Automatically create a Chatter instance based on the configuration.
func NewInstance(c Instance) (chatter.Chatter, error) {
	if !strings.HasPrefix(c.Provider, "provider:") {
		return nil, fmt.Errorf("invalid schema: %s for the provider, provider:{provider}/{capability}/{family} is required", c.Provider)
	}

	switch c.Provider {
	case "provider:mock":
		return &Mock{}, nil

	case "provider:bedrock/embedding/titan":
		return titan.New(c.Model, c.Dimensions, bedrock.WithRegion(c.Region))

	case "provider:bedrock/foundation/converse":
		return converse.New(c.Model, converse.WithRegion(c.Region))

	case "provider:bedrock/foundation/llama":
		return llama.New(c.Model, bedrock.WithRegion(c.Region))

	case "provider:bedrock/foundation/nova":
		return nova.New(c.Model, bedrock.WithRegion(c.Region))

	case "provider:openai/embedding/text2vec":
		return text2vec.New(c.Model, c.Dimensions,
			openai.WithHost(c.Host),
			openai.WithSecret(c.Secret),
			openai.WithHTTP(http.WithClient(curl(c))),
		)

	case "provider:openai/foundation/gpt":
		return gpt.New(c.Model,
			openai.WithHost(c.Host),
			openai.WithSecret(c.Secret),
			openai.WithHTTP(http.WithClient(curl(c))),
		)

	case "provider:google/foundation/gemini":
		return gemini.New(c.Model, gemini.Config{Secret: c.Secret})

	case "provider:google/foundation/imagen":
		return imagen.New(c.Model, imagen.Config{Secret: c.Secret})
	}

	return nil, fmt.Errorf("configuration is not supported: %s, %s", c.Provider, c.Model)
}

func curl(c Instance) *gohttp.Client {
	cli := http.Client()
	cli.Timeout = time.Duration(c.Timeout) * time.Second
	return cli
}
