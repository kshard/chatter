//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio"
	"github.com/kshard/chatter/llm/autoconfig"
)

func main() {
	llm, err := autoconfig.New("chatter")
	if err != nil {
		panic(err)
	}

	assistant := aio.NewLogger(os.Stdout, llm)

	var prompt chatter.Prompt
	prompt.WithTask("Extract keywords from the text: %s", text)

	reply, err := assistant.Prompt(context.Background(),
		prompt.ToSeq(),
		chatter.Temperature(0.9),
		chatter.Quota(512),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n\n==> (in: %d out: %d tokens)\n%s\n",
		assistant.UsedInputTokens(), assistant.UsedReplyTokens(), reply)
}

const text = `
	DynamoDB has truly revolutionize our data management approach. DynamoDB
	scalability is a standout feature. The ability to seamlessly scale based
	on demand has been instrumental in accommodating our growing data
	requirements. The consistent, low-latency performance is commendable.
	Seamless scalability, coupled with lightning-fast performance. Large-scale
	datasets support without any issue. Flexibility and fully managed nature.
	As managed service, the DynamoDB simplifies the operational overhead of
	database management. Features such as automatic backups, security patching,
	and continuous monitoring contribute to a hassle-free experience.
	`
