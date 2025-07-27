//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package batch

import "github.com/fogfish/stream"

type FileSystem struct {
	stream.CreateFS[struct{}]
	bucket string
	role   string
}

func NewFileSystem(bucket, role string, opt ...stream.Option) (*FileSystem, error) {
	fsys, err := stream.NewFS(bucket, opt...)
	if err != nil {
		return nil, err
	}

	return &FileSystem{
		CreateFS: fsys,
		bucket:   bucket,
		role:     role,
	}, nil
}
