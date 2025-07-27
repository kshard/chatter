//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package openai

import (
	"context"
	"fmt"
	"os/user"
	"path/filepath"

	"github.com/fogfish/gurl/v2/http"
	ƒ "github.com/fogfish/gurl/v2/http/recv"
	ø "github.com/fogfish/gurl/v2/http/send"
	"github.com/fogfish/opts"
	"github.com/jdxcode/netrc"
)

//
// Configuration options for OpenAI
//

type Option = opts.Option[Client]

var (
	// Config HTTP stack
	WithHTTP = opts.Use[Client](http.NewStack)

	// Config the host, api.openai.com is default
	WithHost = opts.ForName[Client, string]("host")

	// Config API secret key
	WithSecret = opts.ForName[Client, string]("secret")

	// Set api secret from ~/.netrc file
	WithNetRC = opts.FMap(withNetRC)
)

func withNetRC(h *Client, host string) error {
	if h.secret != "" {
		return nil
	}

	usr, err := user.Current()
	if err != nil {
		return err
	}

	n, err := netrc.Parse(filepath.Join(usr.HomeDir, ".netrc"))
	if err != nil {
		return err
	}

	machine := n.Machine(host)
	if machine == nil {
		return fmt.Errorf("undefined secret for host <%s> at ~/.netrc", host)
	}

	h.secret = machine.Get("password")
	return nil
}

type Client struct {
	http.Stack
	host   string
	path   string
	secret string
}

type Service[A, B any] struct {
	client Client
}

func New[A, B any](path string, opt ...Option) (*Service[A, B], error) {
	c := Client{
		host: "https://api.openai.com",
		path: path,
	}
	if err := opts.Apply(&c, opt); err != nil {
		return nil, err
	}

	if c.Stack == nil {
		c.Stack = http.New()
	}

	return &Service[A, B]{client: c}, nil
}

func (s *Service[A, B]) Invoke(ctx context.Context, input A) (B, error) {
	bag, err := http.IO[B](s.client.WithContext(ctx),
		http.POST(
			ø.URI("%s%s", ø.Authority(s.client.host), ø.Path(s.client.path)),
			ø.Accept.JSON,
			ø.Authorization.Set("Bearer "+s.client.secret),
			ø.ContentType.JSON,
			ø.Send(input),

			ƒ.Status.OK,
			ƒ.ContentType.JSON,
		),
	)
	if err != nil {
		return *new(B), err
	}

	return *bag, nil
}
