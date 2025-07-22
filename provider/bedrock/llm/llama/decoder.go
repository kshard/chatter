package llama

import "github.com/kshard/chatter"

func (decoder decoder) Decode(llama *reply) (*chatter.Reply, error) {
	reply := new(chatter.Reply)

	switch llama.StopReason {
	case "stop":
		reply.Stage = chatter.LLM_RETURN
	case "length":
		reply.Stage = chatter.LLM_INCOMPLETE
	default:
		reply.Stage = chatter.LLM_ERROR
	}

	reply.Usage.InputTokens = llama.UsedPromptTokens
	reply.Usage.ReplyTokens = llama.UsedTextTokens

	reply.Content = []chatter.Content{
		chatter.Text(llama.Text),
	}

	return reply, nil
}
