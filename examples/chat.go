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
	// chat "github.com/kshard/chatter/openai"
	chat "github.com/kshard/chatter/bedrock"
)

func main() {
	session, err := chat.New(chat.WithQuotaTokensInReply(256))
	if err != nil {
		panic(err)
	}

	prompt := chatter.NewPrompt(
		chatter.WithStratum("You are pirate, Captain Blood."),
		chatter.WithContext("\"Captain Blood: His Odyssey\" book by Rafael Sabatini constraints replies."),
	)

	prompt.Inquiry("What we are doing upon arrival on the island of Barbados?")

	prompt, err = session.Send(context.Background(), prompt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("The context is %s\n", prompt.Context)
	fmt.Printf("The model behavior is \"%s\"\n", prompt.Stratum)
	fmt.Println("------")
	for _, m := range prompt.Messages {
		switch m.Role {
		case chatter.INQUIRY:
			fmt.Printf("Q: %s\n\n", m.Content)
		case chatter.CHATTER:
			fmt.Printf("A: %s\n\n", m.Content)
		}
	}
	fmt.Println("------")
	fmt.Printf("Used tokens %d\n", session.ConsumedTokens())
}
