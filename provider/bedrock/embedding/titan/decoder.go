package titan

import "github.com/kshard/chatter"

func (decoder decoder) Decode(titan *reply) (*chatter.Reply, error) {
	reply := new(chatter.Reply)
	reply.Stage = chatter.LLM_RETURN

	reply.Usage.InputTokens = titan.UsedTextTokens

	reply.Content = []chatter.Content{
		chatter.Vector(titan.Vector),
	}

	return reply, nil
}
