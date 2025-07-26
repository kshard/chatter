package gpt

import "github.com/kshard/chatter"

func (decoder decoder) Decode(bag *reply) (*chatter.Reply, error) {
	reply := &chatter.Reply{
		Stage: chatter.LLM_RETURN,
		Content: []chatter.Content{
			chatter.Text(bag.Choices[0].Message.Content),
		},
		Usage: chatter.Usage{
			InputTokens: bag.Usage.PromptTokens,
			ReplyTokens: bag.Usage.OutputTokens,
		},
	}
	return reply, nil
}
