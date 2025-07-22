package titan

import (
	"strings"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/aio/provider"
)

func factory() (provider.Encoder[*input], error) {
	codec := &encoder{
		w:   strings.Builder{},
		req: input{},
	}
	return codec, nil
}

func (codec *encoder) WithTemperature(temp float64)         {}
func (codec *encoder) WithTopP(topP float64)                {}
func (codec *encoder) WithMaxTokens(maxTokens int)          {}
func (codec *encoder) WithStopSequences(sequences []string) {}
func (codec *encoder) WithCommand(cmd chatter.Cmd)          {}

func (codec *encoder) AsStratum(stratum chatter.Stratum) error {
	return nil
}

func (codec *encoder) AsText(text chatter.Text) error {
	codec.w.WriteString(string(text))
	return nil
}

func (codec *encoder) AsPrompt(prompt *chatter.Prompt) error {
	codec.w.WriteString(prompt.String())
	return nil
}

func (codec *encoder) AsAnswer(answer *chatter.Answer) error {
	return nil
}

func (codec *encoder) AsReply(reply *chatter.Reply) error {
	return nil
}

func (codec *encoder) Build() *input {
	codec.req.Text = codec.w.String()
	return &codec.req
}
