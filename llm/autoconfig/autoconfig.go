//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package autoconfig

import (
	"fmt"
	"os/user"
	"path/filepath"

	ø "github.com/fogfish/gurl/v2/http/send"
	"github.com/jdxcode/netrc"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/llm/bedrock"
	"github.com/kshard/chatter/llm/openai"
)

// Configures LLM api automatically from ~/.netrc
func New(host string, model ...string) (chatter.Chatter, error) {
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

	var mid string
	if len(model) > 0 {
		mid = model[0]
	}

	lpd := machine.Get("provider")
	switch lpd {
	case "bedrock":
		return makeBedrock(machine, mid)
	case "openai":
		return makeOpenAI(machine, mid)
	default:
		return nil, fmt.Errorf("provider %s is not supported", lpd)
	}
}

func makeBedrock(conf *netrc.Machine, model string) (chatter.Chatter, error) {
	var (
		llm    bedrock.Option
		family = conf.Get("family")
		region = conf.Get("region")
	)

	if len(model) == 0 {
		model = conf.Get("model")
	}

	switch family {
	case "llama3":
		llm = bedrock.WithLLM(bedrock.Llama3(model))
	default:
		return nil, fmt.Errorf("family %s is not supported", family)
	}

	return bedrock.New(llm, bedrock.WithRegion(region))
}

func makeOpenAI(conf *netrc.Machine, model string) (chatter.Chatter, error) {
	var (
		host   = conf.Get("host")
		secret = conf.Get("secret")
	)

	if len(model) == 0 {
		model = conf.Get("model")
	}

	return openai.New(
		openai.WithHost(ø.Authority(host)),
		openai.WithSecret(secret),
		openai.WithLLM(openai.LLM(model)),
	)
}
