//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

import "context"

// Chatter interface
type Chatter interface {
	ConsumedTokens() int
	Send(context.Context, *Prompt) (*Prompt, error)
}
