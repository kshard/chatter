//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrockbatch

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/bedrock"
	"github.com/fogfish/stream"
)

//-----------------------------------------------------------------------------

// File System API used by the client for S3 I/O
type FileSystem = stream.CreateFS[struct{}]

// AWS Bedrock API used by the client for batch inference
type Bedrock interface {
	CreateModelInvocationJob(ctx context.Context, params *bedrock.CreateModelInvocationJobInput, optFns ...func(*bedrock.Options)) (*bedrock.CreateModelInvocationJobOutput, error)
}
