//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/bedrock"
	"github.com/kshard/chatter/bedrockbatch"
)

func main() {
	assistant, err := bedrockbatch.New(
		bedrockbatch.WithLLM(bedrock.LLAMA3_1_70B_INSTRUCT),
		bedrockbatch.WithBucket("my-bucket"),
		bedrockbatch.WithRole("arn:aws:iam::000000000000:role/my-role"),
	)
	if err != nil {
		panic(err)
	}

	input, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer input.Close()

	job, err := assistant.Prepare()
	if err != nil {
		panic(err)
	}

	var prompt chatter.Prompt
	prompt.WithTask("Extract keywords from the input text.")
	prompt.With(
		chatter.Rules(
			`Strictly adhere to the following requirements when generating a response.
			Do not deviate, ignore, or modify any aspect of them:`,

			"Rank and order keywords according to the relevance.",
			"Return empty list if text is short or no way to extract keywords with high confidence.",
			"Replacing pronouns (e.g., it, he, she, they, this, that) with the full name of the entities they refer to.",
			"Present the results as a list of strings, formatted in JSON. Do not output any explanations.",
		),
	)
	prompt.With(
		chatter.Example{
			Input: `the heat in the street was terrible: and the airlessness, the bustle and the plaster, scaffolding, bricks, and dust all about him, and that special petersburg stench, so familiar to all who are unable to get out of town in summer--all worked painfully upon the young man’s already overwrought nerves.`,
			Reply: `[ "overwrought nerves", "Petersburg", "heat", "airlessness", "dust", "bustle", "scaffolding", "summer", "bricks", "plaster", "street", "young man" ]`,
		},
	)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		prompt.With(
			chatter.Input("Input text:", scanner.Text()),
		)

		_, err := job.Prompt(context.Background(), prompt.ToSeq())
		if err != nil {
			panic(err)
		}
	}

	uid, err := job.Commit()
	if err != nil {
		panic(err)
	}

	fmt.Printf("====>>> %v\n", uid)
}
