package gpt

import (
	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory(model string) func() (provider.Encoder[*input], error) {
	return func() (provider.Encoder[*input], error) {
		return &encoder{req: input{
			Model:    model,
			Messages: []message{},
		},
		}, nil
	}
}

func (codec *encoder) WithInferrer(inferrer provider.Inferrer) {
	codec.req.Temperature = inferrer.Temperature
	codec.req.TopP = inferrer.TopP
	codec.req.MaxTokens = inferrer.MaxTokens
}

func (codec *encoder) WithCommand(cmd chatter.Cmd) {
	// Not supported yet by the library
}

func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	msg := message{Role: "system", Content: string(stratum)}
	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

func (codec *encoder) AsText(text chatter.Text) error {
	msg := message{Role: "user", Content: string(text)}
	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	msg := message{Role: "user", Content: prompt.String()}
	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	// Not supported yet
	return nil
}

func (codec *encoder) AsReply(reply *chatter.Reply) error {
	msg := message{Role: "assistant", Content: reply.String()}
	codec.req.Messages = append(codec.req.Messages, msg)
	return nil
}

func (codec *encoder) Build() *input {
	return &codec.req
}
