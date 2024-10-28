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

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/bedrock"
)

func main() {
	assistant, err := bedrock.New(
		bedrock.WithModel(bedrock.LLAMA3_0_8B_INSTRUCT),
		bedrock.WithQuotaTokensInReply(512),
	)
	if err != nil {
		panic(err)
	}

	var prompt chatter.Prompt
	prompt.WithTask("Extract keywords from the text: %s", text)

	reply, err := assistant.Prompt(context.Background(), &prompt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("==> (%d tokens)\n%s\n", assistant.ConsumedTokens(), reply)
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
