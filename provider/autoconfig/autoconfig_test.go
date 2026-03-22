//
// Copyright (C) 2024 - 2026 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package autoconfig

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/fogfish/it/v2"
	"github.com/kshard/chatter"
)

// ---------------------------------------------------------------------------
// Test fixtures

const yamlSingle = `
mymodel:
  provider: "provider:mock"
  model: "test-model"
`

const yamlMulti = `
model1:
  provider: "provider:mock"
  model: "mock-a"
model2:
  provider: "provider:mock"
  model: "mock-b"
`

const jsonSingle = `{"mymodel": {"provider": "provider:mock", "model": "test-model"}}`

const jsonMulti = `{
  "model1": {"provider": "provider:mock", "model": "mock-a"},
  "model2": {"provider": "provider:mock", "model": "mock-b"}
}`

const netrcSingle = `machine mymodel
  provider provider:mock
  model test-model
`

const netrcMulti = `machine model1
  provider provider:mock
  model mock-a
machine model2
  provider provider:mock
  model mock-b
`

// ---------------------------------------------------------------------------
// FromFile: YAML format

func TestFromFile_YAML(t *testing.T) {
	t.Run("single_instance_yaml", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.yaml": {Data: []byte(yamlSingle)},
		}
		instances, err := FromFile(fsys, "config.yaml")
		it.Then(t).Should(it.Nil(err)).ShouldNot(it.Nil(instances))
	})

	t.Run("yml_extension", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.yml": {Data: []byte(yamlSingle)},
		}
		instances, err := FromFile(fsys, "config.yml")
		it.Then(t).Should(it.Nil(err)).ShouldNot(it.Nil(instances))
	})

	t.Run("multiple_instances_yaml", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.yaml": {Data: []byte(yamlMulti)},
		}
		instances, err := FromFile(fsys, "config.yaml")
		it.Then(t).Should(it.Nil(err)).ShouldNot(it.Nil(instances))

		_, ok1 := instances.Model("model1")
		_, ok2 := instances.Model("model2")
		it.Then(t).Should(
			it.Equal(ok1, true),
			it.Equal(ok2, true),
		)
	})

	t.Run("invalid_yaml_content", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.yaml": {Data: []byte(":\t: invalid yaml ::")},
		}
		instances, err := FromFile(fsys, "config.yaml")
		it.Then(t).Should(it.Error(instances, err))
	})

	t.Run("unsupported_provider_yaml", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.yaml": {Data: []byte(`mymodel:
  provider: "provider:unsupported/unknown"
  model: "test"
`)},
		}
		instances, err := FromFile(fsys, "config.yaml")
		it.Then(t).Should(it.Error(instances, err).Contain("configuration is not supported"))
	})
}

// ---------------------------------------------------------------------------
// FromFile: JSON format

func TestFromFile_JSON(t *testing.T) {
	t.Run("single_instance_json", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.json": {Data: []byte(jsonSingle)},
		}
		instances, err := FromFile(fsys, "config.json")
		it.Then(t).Should(it.Nil(err)).ShouldNot(it.Nil(instances))
	})

	t.Run("multiple_instances_json", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.json": {Data: []byte(jsonMulti)},
		}
		instances, err := FromFile(fsys, "config.json")
		it.Then(t).Should(it.Nil(err)).ShouldNot(it.Nil(instances))

		_, ok1 := instances.Model("model1")
		_, ok2 := instances.Model("model2")
		it.Then(t).Should(
			it.Equal(ok1, true),
			it.Equal(ok2, true),
		)
	})

	t.Run("invalid_json_content", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.json": {Data: []byte("{ invalid json }")},
		}
		instances, err := FromFile(fsys, "config.json")
		it.Then(t).Should(it.Error(instances, err))
	})

	t.Run("unsupported_provider_json", func(t *testing.T) {
		fsys := fstest.MapFS{
			"config.json": {Data: []byte(`{"mymodel": {"provider": "provider:unsupported/unknown", "model": "test"}}`)},
		}
		instances, err := FromFile(fsys, "config.json")
		it.Then(t).Should(it.Error(instances, err).Contain("configuration is not supported"))
	})
}

// ---------------------------------------------------------------------------
// FromFile: netrc format (default for all other extensions)

func TestFromFile_NetRC(t *testing.T) {
	t.Run("single_machine_netrc", func(t *testing.T) {
		fsys := fstest.MapFS{
			".config": {Data: []byte(netrcSingle)},
		}
		instances, err := FromFile(fsys, ".config")
		it.Then(t).Should(it.Nil(err)).ShouldNot(it.Nil(instances))

		_, ok := instances.Model("mymodel")
		it.Then(t).Should(it.Equal(ok, true))
	})

	t.Run("multiple_machines_netrc", func(t *testing.T) {
		fsys := fstest.MapFS{
			".netrc": {Data: []byte(netrcMulti)},
		}
		instances, err := FromFile(fsys, ".netrc")
		it.Then(t).Should(it.Nil(err)).ShouldNot(it.Nil(instances))

		_, ok1 := instances.Model("model1")
		_, ok2 := instances.Model("model2")
		it.Then(t).Should(
			it.Equal(ok1, true),
			it.Equal(ok2, true),
		)
	})

	t.Run("unsupported_provider_netrc", func(t *testing.T) {
		fsys := fstest.MapFS{
			".netrc": {Data: []byte(`machine mymodel
  provider provider:unsupported/unknown
  model test-model
`)},
		}
		instances, err := FromFile(fsys, ".netrc")
		it.Then(t).Should(it.Error(instances, err).Contain("configuration is not supported"))
	})
}

// ---------------------------------------------------------------------------
// FromFile: I/O errors

func TestFromFile_NotFound(t *testing.T) {
	t.Run("file_not_found_yaml", func(t *testing.T) {
		fsys := fstest.MapFS{}
		instances, err := FromFile(fsys, "nonexistent.yaml")
		it.Then(t).Should(it.Error(instances, err))
	})

	t.Run("file_not_found_json", func(t *testing.T) {
		fsys := fstest.MapFS{}
		instances, err := FromFile(fsys, "nonexistent.json")
		it.Then(t).Should(it.Error(instances, err))
	})

	t.Run("file_not_found_netrc", func(t *testing.T) {
		fsys := fstest.MapFS{}
		instances, err := FromFile(fsys, ".missing")
		it.Then(t).Should(it.Error(instances, err))
	})
}

// ---------------------------------------------------------------------------
// Instances.Get

func TestInstances_Get(t *testing.T) {
	fsys := fstest.MapFS{
		"config.yaml": {Data: []byte(yamlSingle)},
	}
	instances, err := FromFile(fsys, "config.yaml")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("returns_mock_llm_for_known_key", func(t *testing.T) {
		llm, ok := instances.Model("mymodel")
		it.Then(t).
			Should(it.Equal(ok, true)).
			ShouldNot(it.Nil(llm))
	})

	t.Run("returns_false_for_unknown_key", func(t *testing.T) {
		_, ok := instances.Model("nonexistent")
		it.Then(t).Should(it.Equal(ok, false))
	})
}

// ---------------------------------------------------------------------------
// Instances.Usage

func TestInstances_Usage(t *testing.T) {
	fsys := fstest.MapFS{
		"config.yaml": {Data: []byte(yamlMulti)},
	}
	instances, err := FromFile(fsys, "config.yaml")
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	t.Run("zero_usage_before_any_prompt", func(t *testing.T) {
		total, breakdown := instances.Usage()
		it.Then(t).Should(
			it.Equal(total.InputTokens, 0),
			it.Equal(total.ReplyTokens, 0),
			it.Equal(len(breakdown), 0),
		)
	})

	t.Run("tracks_usage_after_prompt", func(t *testing.T) {
		llm, _ := instances.Model("model1")
		msg := chatter.Stratum("hello world")
		_, err := llm.Prompt(context.Background(), []chatter.Message{msg})
		if err != nil {
			t.Fatalf("prompt failed: %v", err)
		}

		total, breakdown := instances.Usage()
		it.Then(t).Should(
			it.Equal(len(breakdown), 1),
		).ShouldNot(
			it.Equal(total.InputTokens, 0),
		)
	})
}
