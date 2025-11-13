//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package gemini

import (
	"fmt"

	"github.com/kshard/chatter"
	"google.golang.org/genai"
)

//------------------------------------------------------------------------------

type decoder struct{}

func (decoder decoder) Decode(bag *genai.GenerateContentResponse) (*chatter.Reply, error) {
	if len(bag.Candidates) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	content := []chatter.Content{}
	for _, part := range bag.Candidates[0].Content.Parts {
		if part.Text != "" {
			content = append(content, chatter.Text(part.Text))
		} else if part.InlineData != nil {
			content = append(content, &chatter.Binary{
				Data: part.InlineData.Data,
				Type: part.InlineData.MIMEType,
			})
		}
	}

	reply := &chatter.Reply{
		Stage:   chatter.LLM_RETURN,
		Content: content,
		Usage: chatter.Usage{
			InputTokens: int(bag.UsageMetadata.ToolUsePromptTokenCount),
			ReplyTokens: int(bag.UsageMetadata.CandidatesTokenCount),
		},
	}
	return reply, nil
}
