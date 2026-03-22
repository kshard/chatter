//
// Copyright (C) 2024 - 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package autoconfig

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/jdxcode/netrc"
	"github.com/kshard/chatter"
)

// LLM instances for the application
type Instances struct {
	Spec map[string]Instance
	llms map[string]chatter.Chatter
}

func (i Instances) Model(name string) (chatter.Chatter, bool) {
	if llm, ok := i.llms[name]; ok {
		return llm, true
	}

	return nil, false
}

func (i *Instances) Build() error {
	i.llms = make(map[string]chatter.Chatter)

	for name, spec := range i.Spec {
		if spec.Provider == "" {
			continue
		}

		llm, err := NewInstance(spec)
		if err != nil {
			return fmt.Errorf("failed to build instance %s: %w", name, err)
		}

		i.llms[name] = llm
	}

	return nil
}

func (i Instances) Usage() (chatter.Usage, map[string]chatter.Usage) {
	total := chatter.Usage{}
	usage := make(map[string]chatter.Usage)
	for name, llm := range i.llms {
		u := llm.Usage()
		if u.InputTokens == 0 && u.ReplyTokens == 0 {
			continue
		}
		usage[name] = u
		total.InputTokens += u.InputTokens
		total.ReplyTokens += u.ReplyTokens
	}
	return total, usage
}

// Configures from file on local file system, panic if the file is not found or the configuration is invalid.
func FromConfig(path string) (*Instances, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(usr.HomeDir, path[1:])
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	return FromFile(os.DirFS("/"), path[1:])
}

// Configures from file on local file system, panic if the file is not found or the configuration is invalid.
func MustFromConfig(path string) *Instances {
	cfg, err := FromConfig(path)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration from %s: %v", path, err))
	}

	return cfg
}

// Configures from file on file system
func FromFile(fsys fs.FS, path string) (*Instances, error) {
	r, err := fsys.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	switch {
	case strings.HasSuffix(path, ".yaml"), strings.HasSuffix(path, ".yml"):
		cfg, err := FromYAML(r)
		if err != nil {
			return nil, err
		}

		return cfg, nil

	case strings.HasSuffix(path, ".json"):
		cfg, err := FromJSON(r)
		if err != nil {
			return nil, err
		}

		return cfg, nil

	default:
		cfg, err := FromNetRC(r)
		if err != nil {
			return nil, err
		}

		return cfg, nil
	}
}

func MustFromFile(fsys fs.FS, path string) *Instances {
	cfg, err := FromFile(fsys, path)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration from %s: %v", path, err))
	}

	return cfg
}

// Configures llm instances from YAML syntax
func FromYAML(r io.Reader) (*Instances, error) {
	var cfg Instances

	if err := yaml.NewDecoder(r).Decode(&cfg.Spec); err != nil {
		return nil, err
	}

	if err := cfg.Build(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustFromYAML(r io.Reader) *Instances {
	cfg, err := FromYAML(r)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration from YAML: %v", err))
	}

	return cfg
}

// Configures llm instances from JSON syntax
func FromJSON(r io.Reader) (*Instances, error) {
	var cfg Instances

	if err := json.NewDecoder(r).Decode(&cfg.Spec); err != nil {
		return nil, err
	}

	if err := cfg.Build(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustFromJSON(r io.Reader) *Instances {
	cfg, err := FromJSON(r)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration from JSON: %v", err))
	}

	return cfg
}

// Configures llm instances from netrc syntax
func FromNetRC(r io.Reader) (*Instances, error) {
	cfg := Instances{Spec: make(map[string]Instance)}

	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	n, err := netrc.ParseString(string(b))
	if err != nil {
		return nil, err
	}

	for _, machine := range n.Machines() {
		timeout := 300
		if val := machine.Get("timeout"); len(val) != 0 {
			if sec, err := strconv.Atoi(val); err == nil {
				timeout = sec
			}
		}

		dimensions := 0
		if val := machine.Get("dimensions"); len(val) != 0 {
			if dim, err := strconv.Atoi(val); err == nil {
				dimensions = dim
			}
		}

		cfg.Spec[machine.Name] = Instance{
			Provider:   machine.Get("provider"),
			Model:      machine.Get("model"),
			Region:     machine.Get("region"),
			Host:       machine.Get("host"),
			Secret:     machine.Get("secret"),
			Timeout:    timeout,
			Dimensions: dimensions,
		}
	}

	if err := cfg.Build(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func MustFromNetRC(r io.Reader) *Instances {
	cfg, err := FromNetRC(r)
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration from netrc: %v", err))
	}

	return cfg
}
