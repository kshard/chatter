//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
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
	"time"

	"github.com/aws/aws-sdk-go-v2/service/bedrock/types"
	"github.com/fogfish/stream"
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/provider/bedrock/batch"
	"github.com/kshard/chatter/provider/bedrock/foundation/nova"
)

// Amazon Nova Lite only supports batch inference as a single-region model
// in a handful of regions (notably not us-west-2, this library's default).
// See https://docs.aws.amazon.com/bedrock/latest/userguide/batch-inference-supported.html
const model = "amazon.nova-lite-v1:0"

// Extracts keywords from each line of the input file using AWS Bedrock
// Batch Inference: every line is enqueued as a separate record, the whole
// batch is submitted as a single job, and results are correlated back to
// the line that produced them once the job completes.
//
// The region must be one where Nova Lite supports batch inference
// (e.g. us-east-1, me-central-1) and must match the region the S3 bucket
// and IAM role from ../iac were deployed to.
//
//	go run . <region> <s3-bucket> <batch-role-arn> <input-file>
func main() {
	if len(os.Args) != 5 {
		fmt.Fprintln(os.Stderr, "usage: batch <region> <s3-bucket> <batch-role-arn> <input-file>")
		os.Exit(1)
	}
	region, bucket, role, path := os.Args[1], os.Args[2], os.Args[3], os.Args[4]

	fsys, err := batch.NewFileSystem(bucket, role, stream.WithRegion(region))
	if err != nil {
		panic(err)
	}

	assistant, err := nova.NewBatch(fsys, model, batch.WithRegion(region))
	if err != nil {
		panic(err)
	}

	job, err := assistant.Prepare()
	if err != nil {
		panic(err)
	}

	input, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer input.Close()

	// Job.Prompt returns a record identifier instead of a completion:
	// batch inference is asynchronous, results only become available once
	// the job finishes. Keep the identifier around to match it against the
	// line that produced it once Fetch decodes the results.
	lines := make(map[string]string)

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		line := scanner.Text()

		id, err := job.Prompt(context.Background(), keywordsPrompt(line))
		if err != nil {
			panic(err)
		}
		lines[id] = line
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}

	arn, err := job.Commit()
	if err != nil {
		panic(err)
	}
	fmt.Printf("submitted batch job: %s\n", arn)

	for {
		status, err := job.Status(context.Background())
		if err != nil {
			panic(err)
		}
		fmt.Printf("status: %s\n", status)

		switch status {
		case types.ModelInvocationJobStatusCompleted, types.ModelInvocationJobStatusPartiallyCompleted:
			report(job, lines)
			return

		case types.ModelInvocationJobStatusFailed,
			types.ModelInvocationJobStatusStopped,
			types.ModelInvocationJobStatusExpired:
			panic(fmt.Sprintf("batch job did not complete: %s", status))
		}

		// Batch jobs take minutes to hours to complete; poll infrequently.
		time.Sleep(30 * time.Second)
	}
}

func report[A, B any](job *batch.Job[A, B], lines map[string]string) {
	results, err := job.Fetch(context.Background())
	if err != nil {
		panic(err)
	}

	for _, result := range results {
		line := lines[result.ID]

		if result.Err != nil {
			fmt.Printf("%s\t[error] %s\n", line, result.Err)
			continue
		}
		fmt.Printf("%s\t=> %s\n", line, result.Reply)
	}
}

func keywordsPrompt(line string) []chatter.Message {
	var prompt chatter.Prompt
	prompt.WithTask("Extract keywords from the input text.")
	prompt.WithRules(
		`Strictly adhere to the following requirements when generating a response.
		Do not deviate, ignore, or modify any aspect of them:`,

		"Rank and order keywords according to the relevance.",
		"Return empty list if text is short or no way to extract keywords with high confidence.",
		"Replacing pronouns (e.g., it, he, she, they, this, that) with the full name of the entities they refer to.",
		"Present the results as a list of strings, formatted in JSON. Do not output any explanations.",
	)
	prompt.WithExample(
		`the heat in the street was terrible: and the airlessness, the bustle and the plaster, scaffolding, bricks, and dust all about him, and that special petersburg stench, so familiar to all who are unable to get out of town in summer--all worked painfully upon the young man's already overwrought nerves.`,
		`[ "overwrought nerves", "Petersburg", "heat", "airlessness", "dust", "bustle", "scaffolding", "summer", "bricks", "plaster", "street", "young man" ]`,
	)
	prompt.WithInput("Input text:", line)

	return prompt.ToSeq()
}
