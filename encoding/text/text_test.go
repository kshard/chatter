//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package text_test

import (
	"strings"
	"testing"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter/encoding/text"
)

func TestEncoder(t *testing.T) {
	var sb strings.Builder

	codec, err := text.NewEncoder(&sb, "Bot: ", "User: ", "Role")
	it.Then(t).Should(it.Nil(err))

	err = codec.Write("request")
	it.Then(t).Should(it.Nil(err))

	err = codec.Write("response")
	it.Then(t).Should(it.Nil(err))

	it.Then(t).Should(
		it.Equal(sb.String(), "Role.\n\nUser: request\n\nBot: response\n\n"),
	)
}
