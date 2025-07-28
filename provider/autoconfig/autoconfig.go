//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package autoconfig

import (
	"fmt"
	gohttp "net/http"
	"os/user"
	"path/filepath"

	"strconv"
	"time"

	"github.com/fogfish/curie/v2"
	"github.com/fogfish/gurl/v2/http"
	"github.com/jdxcode/netrc"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/provider/bedrock"
	"github.com/kshard/chatter/provider/bedrock/embedding/titan"
	"github.com/kshard/chatter/provider/bedrock/foundation/converse"
	"github.com/kshard/chatter/provider/bedrock/foundation/llama"
	"github.com/kshard/chatter/provider/bedrock/foundation/nova"
	"github.com/kshard/chatter/provider/openai"
	"github.com/kshard/chatter/provider/openai/embedding/text2vec"
	"github.com/kshard/chatter/provider/openai/foundation/gpt"
)

// Config storage for LLM provider configuration
type Config interface {
	Provider() curie.IRI
	Model() string
	Get(string) string
}

// Configuration parameters/options used by autoconfig
const (
	opt_provider   = "provider"
	opt_model      = "model"
	opt_region     = "region"     // used by Bedrock providers
	opt_host       = "host"       // used by OpenAI providers
	opt_secret     = "secret"     // used by OpenAI providers
	opt_timeout    = "timeout"    // used by OpenAI providers
	opt_dimensions = "dimensions" // used by embedding families
)

// Automatically create a Chatter instance based on the configuration.
func New(c Config, model ...string) (chatter.Chatter, error) {
	provider := c.Provider()
	if curie.Schema(provider) != opt_provider {
		return nil, fmt.Errorf("invalid schema: %s for the provider, provider:{provider}/{capability}/{family} is required", provider)
	}

	fmID := c.Model()
	if len(model) > 0 {
		fmID = model[0]
	}

	switch provider {
	case "provider:bedrock/embedding/titan":
		dim, err := strconv.Atoi(c.Get(opt_dimensions))
		if err != nil {
			return nil, fmt.Errorf("invalid config, dimensions: %w", err)
		}
		return titan.New(fmID, dim, bedrock.WithRegion(c.Get(opt_region)))

	case "provider:bedrock/foundation/converse":
		return converse.New(fmID, converse.WithRegion(c.Get(opt_region)))

	case "provider:bedrock/foundation/llama":
		return llama.New(fmID, bedrock.WithRegion(c.Get(opt_region)))

	case "provider:bedrock/foundation/nova":
		return nova.New(fmID, bedrock.WithRegion(c.Get(opt_region)))

	case "provider:openai/embedding/text2vec":
		dim, err := strconv.Atoi(c.Get(opt_dimensions))
		if err != nil {
			return nil, fmt.Errorf("invalid config, dimensions: %w", err)
		}

		return text2vec.New(fmID, dim,
			openai.WithHost(c.Get(opt_host)),
			openai.WithSecret(c.Get(opt_secret)),
			openai.WithHTTP(http.WithClient(curl(c))),
		)

	case "provider:openai/foundation/gpt":
		return gpt.New(fmID,
			openai.WithHost(c.Get(opt_host)),
			openai.WithSecret(c.Get(opt_secret)),
			openai.WithHTTP(http.WithClient(curl(c))),
		)
	}

	return nil, fmt.Errorf("configuration is not supported: %s", c.Model())
}

func curl(c Config) *gohttp.Client {
	cli := http.Client()
	cli.Timeout = 2 * time.Minute

	if seconds := c.Get(opt_timeout); len(seconds) != 0 {
		if sec, err := strconv.Atoi(seconds); err == nil {
			cli.Timeout = time.Duration(sec) * time.Second
		}
	}

	return cli
}

// configuration provider from ~/.netrc file
type netconfig struct {
	machine *netrc.Machine
}

var _ Config = netconfig{}

// Configure LLM interface using ~/.netrc configuration
func FromNetRC(host string, model ...string) (chatter.Chatter, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		return nil, err
	}

	machine := n.Machine(host)
	if machine == nil {
		return nil, fmt.Errorf("undefined config for <%s> at ~/.netrc", host)
	}

	return New(netconfig{machine: machine}, model...)
}

func (c netconfig) Provider() curie.IRI {
	return curie.IRI(c.machine.Get(opt_provider))
}

func (c netconfig) Model() string {
	return c.machine.Get(opt_model)
}

func (c netconfig) Get(key string) string {
	return c.machine.Get(key)
}
