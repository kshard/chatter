//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrock

import (
	"context"
	"encoding"

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/kshard/chatter"
)

// AWS Bedrock Runtime API
type Bedrock interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

// Bedrock Foundational LLMs
type LLM interface {
	ID() string

	// Encode prompt to bytes:
	// - encoding prompt as prompt markup supported by LLM
	// - encoding prompt to envelop supported by bedrock
	Encode(encoding.TextMarshaler, *chatter.Options) ([]byte, error)

	// Decode LLM's reply into pure text
	Decode([]byte) (Reply, error)
}

type Reply struct {
	Text            string
	UsedInputTokens int
	UsedReplyTokens int
}
