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

	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/kshard/chatter"
)

// AWS Bedrock Runtime API
type Bedrock interface {
	InvokeModel(ctx context.Context, params *bedrockruntime.InvokeModelInput, optFns ...func(*bedrockruntime.Options)) (*bedrockruntime.InvokeModelOutput, error)
}

type LLM chatter.LLM
