//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package chatter

// Prompt data type consisting of context and bag of exchange messages.
type Prompt struct {
	// Ground level constrain of the model behavior.
	// The latin meaning "something that has been laid down".
	// Think about it as a cornerstone of the model behavior.
	// "Act as <stratum>" ...
	Stratum string `json:"stratum,omitempty"`

	// Desired context of prompt and reply.
	Context string `json:"context,omitempty"`

	// Sequence of inquiries and replies
	Messages []Message `json:"messages,omitempty"`
}

// Message of the prompt
type Message struct {
	Role    Role   `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

type Role int

const (
	INQUIRY Role = iota + 1
	CHATTER
)

// Prompt setter interface
type Setter func(prompt *Prompt)

// Set up stratum for newly created prompt
func WithStratum(content string) Setter {
	return func(prompt *Prompt) {
		prompt.Stratum = content
	}
}

// Set up context for newly created prompt
func WithContext(content string) Setter {
	return func(prompt *Prompt) {
		prompt.Context = content
	}
}

// Create new prompt object
func NewPrompt(opts ...Setter) *Prompt {
	prompt := &Prompt{}
	for _, opt := range opts {
		opt(prompt)
	}
	return prompt
}

// Inject inquiry message into prompts
func (prompt *Prompt) Inquiry(content string) *Prompt {
	prompt.Messages = append(prompt.Messages,
		Message{Role: INQUIRY, Content: content},
	)
	return prompt
}

// Inject chatter/reply/generated message into prompts
func (prompt *Prompt) Chatter(content string) *Prompt {
	prompt.Messages = append(prompt.Messages,
		Message{Role: CHATTER, Content: content},
	)
	return prompt
}

// Retrieve the latest reply
func (prompt *Prompt) Reply() string {
	for i := len(prompt.Messages) - 1; i >= 0; i-- {
		msg := prompt.Messages[i]

		if msg.Role == CHATTER {
			return msg.Content
		}
	}

	return ""
}
