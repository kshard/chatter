//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package llama3_test

import (
	"strings"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter/encoding/llama3"
)

func TestEncoder(t *testing.T) {
	var sb strings.Builder

	codec, err := llama3.NewEncoder(&sb, "Role")
	it.Then(t).Should(it.Nil(err))

	err = codec.Write("request")
	it.Then(t).Should(it.Nil(err))

	err = codec.Write("response")
	it.Then(t).Should(it.Nil(err))

	it.Then(t).Should(
		it.Equal(sb.String(), "<|begin_of_text|>\n<|start_header_id|>system<|end_header_id|>\nRole\n<|eot_id|>\n\n<|start_header_id|>user<|end_header_id|>\nrequest\n<|eot_id|>\n\n<|start_header_id|>assistant<|end_header_id|>\nresponse\n<|eot_id|>\n"),
	)
}
