//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package aio

import (
	"context"
	"fmt"

	"github.com/kshard/chatter"
)

// Quoting strategy for LLM I/O
type Quota struct {
	chatter.Chatter
	maxEpoch int
	epoch    int

	maxUsage chatter.Usage
	usage    chatter.Usage
}

var _ chatter.Chatter = (*Quota)(nil)

func NewQuota(maxEpoch int, maxUsage chatter.Usage, chatter chatter.Chatter) *Quota {
	return &Quota{
		Chatter:  chatter,
		maxEpoch: maxEpoch,
		epoch:    0,
		maxUsage: maxUsage,
	}
}

func (q *Quota) ResetQuota() {
	q.epoch = 0
	q.usage.InputTokens = 0
	q.usage.ReplyTokens = 0
}

func (q *Quota) Prompt(ctx context.Context, prompt []chatter.Message, opts ...chatter.Opt) (*chatter.Reply, error) {
	if q.maxEpoch > 0 {
		if q.epoch >= q.maxEpoch {
			return nil, fmt.Errorf("execution aborted, %d epoch is exceeded the quota", q.epoch)
		}
		q.epoch++
	}

	if q.maxUsage.InputTokens > 0 {
		if q.usage.InputTokens >= q.maxUsage.InputTokens {
			return nil, fmt.Errorf("execution aborted, %d input tokens is exceeded the quota", q.usage.InputTokens)
		}
	}

	if q.maxUsage.ReplyTokens > 0 {
		if q.usage.ReplyTokens >= q.maxUsage.ReplyTokens {
			return nil, fmt.Errorf("execution aborted, %d reply tokens is exceeded the quota", q.usage.InputTokens)
		}
	}

	reply, err := q.Chatter.Prompt(ctx, prompt, opts...)
	if err != nil {
		return nil, err
	}

	q.usage.InputTokens += reply.Usage.InputTokens
	q.usage.ReplyTokens += reply.Usage.ReplyTokens

	return reply, nil
}
