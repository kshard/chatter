//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package converse

import (
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/fogfish/opts"
	"github.com/kshard/chatter/aio/provider"
)

type encoder struct {
	req *bedrockruntime.ConverseInput
}

type decoder struct{}

type Converse = provider.Provider[*bedrockruntime.ConverseInput, *bedrockruntime.ConverseOutput]

func New(model string, opt ...Option) (*Converse, error) {
	c := &Service{}

	if err := opts.Apply(c, opt); err != nil {
		return nil, err
	}

	if c.api == nil {
		if err := optsFromRegion(c, defaultRegion); err != nil {
			return nil, err
		}
	}

	return provider.New(factory(model, c.registry), decoder{}, c), nil
}
