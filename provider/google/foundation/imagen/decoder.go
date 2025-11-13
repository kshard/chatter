//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package imagen

import (
	"fmt"

	"github.com/kshard/chatter"
	"google.golang.org/genai"
)

//------------------------------------------------------------------------------

type decoder struct{}

func (decoder decoder) Decode(bag *genai.GenerateImagesResponse) (*chatter.Reply, error) {
	if len(bag.GeneratedImages) == 0 {
		return nil, fmt.Errorf("no content generated")
	}

	content := []chatter.Content{}
	for _, part := range bag.GeneratedImages {
		content = append(content,
			&chatter.Binary{
				Data: part.Image.ImageBytes,
				Type: part.Image.MIMEType,
			})
	}

	reply := &chatter.Reply{
		Stage:   chatter.LLM_RETURN,
		Content: content,
		Usage:   chatter.Usage{},
	}
	return reply, nil
}
