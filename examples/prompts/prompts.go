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
	"strings"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/llm/autoconfig"
)

// There are three reviewers who provided the feedback about AWS DynamoDB
var review = []string{
	`DynamoDB has truly revolutionize our data management approach. DynamoDB
	scalability is a standout feature. The ability to seamlessly scale based
	on demand has been instrumental in accommodating our growing data
	requirements. The consistent, low-latency performance is commendable.
	Seamless scalability, coupled with lightning-fast performance. Large-scale
	datasets support without any issue. Flexibility and fully managed nature.
	As managed service, the DynamoDB simplifies the operational overhead of
	database management. Features such as automatic backups, security patching,
	and continuous monitoring contribute to a hassle-free experience.`,

	`AWS DynamoDB is a super fast key-value pair DB store that is highly scalable
	and highly flexible. It provides read latency of single digit milliseconds
	and provides great integrations with other AWS services like AWS Lambda,
	AWS S3 via DynamoDB streams or Kinesis streams. It has helped make
	out microservices become highly scalable. It is easy to use and highly
	scalable and its SDK are present in all most all modern languages.`,

	`DynamoDB has been a great DB solution for my team and I. The setup process
	is very easy to get started and iterate fast. We utilized it for direct
	lookups of pieces of data and it performs extremely well for that.
	It's a great NoSQL database option. The cost can be high for inserting
	large amounts of data and large-scale applications due to their read/write
	credit system. Limited query flexibility compared to traditional
	SQL databases. There is a learning curve for optimizing data models and
	queries.`,
}

func main() {
	assistant, err := autoconfig.New("chatter")
	if err != nil {
		panic(err)
	}

	prompts := []chatter.Prompt{
		level0(),
		level1(),
		level2(),
		level3(),
		level4(),
		level5(),
		level6(),
	}

	for i, prompt := range prompts {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("## Level: %d\n", i))
		sb.WriteString("### Question\n")
		sb.WriteString(prompt.String())
		sb.WriteString("\n")

		reply, err := assistant.Prompt(context.Background(), prompt.ToSeq())
		if err != nil {
			panic(err)
		}
		sb.WriteString("### Answer\n\n")
		sb.WriteString(reply.String())
		sb.WriteString("\n\n")

		fmt.Println(sb.String())
	}
}

func level0() (prompt chatter.Prompt) {
	prompt.With(chatter.Input("", review...))
	return
}

func level1() (prompt chatter.Prompt) {
	prompt.WithTask("Prepare a review by summarizing the reviewer comments.")
	prompt.With(chatter.Input("", review...))
	return
}

func level2() (prompt chatter.Prompt) {
	prompt.WithTask(`
		Prepare a review by summarizing the following reviewer comments.
		The final output should highlight the core features of the technology,
		common strengths/weaknesses mentioned by multiple reviewers, suggestions
		for improvement.
	`)

	prompt.With(
		chatter.Input("The text for review is below:", review...),
	)
	return
}

func level3() (prompt chatter.Prompt) {
	prompt.WithTask(`
		Prepare a review by answering the following questions from the reviewer
		comments.
	`)

	prompt.With(chatter.Rules("",
		`Based on the reviewer's comments, what are the core contributions made
		by the technology?`,

		`What are the common strengths of this technology, as mentioned by
		multiple reviewers?`,

		`What are the common weaknesses of this technology, as highlighted by
		multiple reviewers?`,

		`What suggestions would you provide for improving this technology?`,
	))

	prompt.With(
		chatter.Input("The text for review is below:", review...),
	)

	return
}

func level4() (prompt chatter.Prompt) {
	prompt = level3()

	prompt.With(
		chatter.Guide(`
			An output should highlight major strengths and issues mentioned by multiple
			reviewers, be less than 400 words in length, the response should be
			in English only.
		`),
	)
	return
}

func level5() (prompt chatter.Prompt) {
	prompt = level4()

	prompt.With(
		chatter.Guide(`Use additional context to answer given questions.`),
	)

	prompt.With(
		chatter.Context(
			"Below are additional context relevant to your goal task.",
			"the traditional data normalization techniques would not work with this database.",
			"the overall data design is based on understanding access patterns.",
			"the database is not designed for supporting SQL-like access.",
			"the first step in designing your DynamoDB application is to identify the specific query patterns that the system must satisfy.",
		),
	)

	return
}

func level6() (prompt chatter.Prompt) {
	prompt = level5()

	prompt.With(
		chatter.Guide(`
			Justify your response in detail by explaining why you made the choices
			you actually made.
		`),
	)

	return
}
