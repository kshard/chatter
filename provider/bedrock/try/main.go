package main

import (
	"context"
	"fmt"

	"github.com/kshard/chatter"
	// "github.com/kshard/chatter/provider/bedrock"
	// "github.com/kshard/chatter/provider/bedrock/embedding/titan"
	"github.com/kshard/chatter/provider/bedrock/converse"
)

func main() {
	// m, _ := llama.New("us.meta.llama3-3-70b-instruct-v1:0")
	// m, _ := titan.New("amazon.titan-embed-text-v2:0",
	// 	bedrock.WithRegion("us-east-1"),
	// )
	m, _ := converse.New("us.anthropic.claude-3-7-sonnet-20250219-v1:0")

	var prompt chatter.Prompt
	prompt.WithTask("Enumerate rainbow colors.")

	r, err := m.Prompt(context.Background(), prompt.ToSeq())
	fmt.Printf("==> %v\n", err)
	fmt.Printf("==> %v\n", r)
	// fmt.Printf("==> %+v\n", r.Content[0].(chatter.Vector))
}
