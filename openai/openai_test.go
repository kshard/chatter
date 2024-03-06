//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kshard/chatter"
)

func TestInquiry(t *testing.T) {
	ts := mock()
	defer ts.Close()

	c, err := New(WithHost(ts.URL), WithSecret("none"))
	if err != nil {
		t.Fatalf("unable to connect to api %s", err)
	}

	for expected, prompt := range map[string]*chatter.Prompt{
		"|user|test|": chatter.NewPrompt().
			Inquiry("test"),

		"|system|stratum|user|test|": chatter.NewPrompt(
			chatter.WithStratum("stratum"),
		).Inquiry("test"),

		"|system|stratum|system|Context: context|user|test|": chatter.NewPrompt(
			chatter.WithStratum("stratum"),
			chatter.WithContext("context"),
		).Inquiry("test"),
	} {
		prompt, err = c.Send(context.Background(), prompt)
		if err != nil {
			t.Errorf("request failed %s", err)
		}

		if prompt.Reply() != expected {
			t.Errorf("unexpected reply %s", prompt.Reply())
		}
	}

}

func mock() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/chat/completions" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			b, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var req modelInquery
			if err := json.Unmarshal(b, &req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			sb := strings.Builder{}
			for _, m := range req.Messages {
				sb.WriteRune('|')
				sb.WriteString(m.Role)
				sb.WriteRune('|')
				sb.WriteString(m.Content)
			}
			sb.WriteRune('|')

			out := &modelChatter{}
			out.Choices = []choice{{message{Content: sb.String()}}}

			d, err := json.Marshal(out)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Add("Content-Type", "application/json")
			w.Write(d)
		}),
	)
}
