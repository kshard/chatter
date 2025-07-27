//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package autoconfig

import (
	"testing"
	"time"

	"github.com/fogfish/curie/v2"
	"github.com/fogfish/it/v2"
)

// Mock implementation of Config interface for testing
type mockConfig struct {
	provider curie.IRI
	model    string
	data     map[string]string
}

func newMockConfig(provider, model string) *mockConfig {
	return &mockConfig{
		provider: curie.IRI(provider),
		model:    model,
		data:     make(map[string]string),
	}
}

func (m *mockConfig) Provider() curie.IRI {
	return m.provider
}

func (m *mockConfig) Model() string {
	return m.model
}

func (m *mockConfig) Get(key string) string {
	return m.data[key]
}

func (m *mockConfig) Set(key, value string) {
	m.data[key] = value
}

func TestNewValidConfigurations(t *testing.T) {
	t.Run("bedrock_embedding_titan", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/embedding/titan", "titan-embed-text-v1")
		config.Set(opt_dimensions, "512")
		config.Set(opt_region, "us-east-1")

		chatter, err := New(config)

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("bedrock_foundation_converse", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/foundation/converse", "claude-3-sonnet")
		config.Set(opt_region, "us-west-2")

		chatter, err := New(config)

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("bedrock_foundation_llama", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/foundation/llama", "llama3-8b-instruct")
		config.Set(opt_region, "eu-west-1")

		chatter, err := New(config)

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("bedrock_foundation_nova", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/foundation/nova", "nova-micro")
		config.Set(opt_region, "us-east-1")

		chatter, err := New(config)

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("openai_embedding_text2vec", func(t *testing.T) {
		config := newMockConfig("provider:openai/embedding/text2vec", "text-embedding-3-small")
		config.Set(opt_dimensions, "1536")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		chatter, err := New(config)

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("openai_foundation_gpt", func(t *testing.T) {
		config := newMockConfig("provider:openai/foundation/gpt", "gpt-4")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		chatter, err := New(config)

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})
}

func TestNewWithModelOverride(t *testing.T) {
	t.Run("override_model_id", func(t *testing.T) {
		config := newMockConfig("provider:openai/foundation/gpt", "gpt-3.5-turbo")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		// Override with different model
		chatter, err := New(config, "gpt-4-turbo")

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("multiple_model_override_uses_first", func(t *testing.T) {
		config := newMockConfig("provider:openai/foundation/gpt", "gpt-3.5-turbo")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		// Pass multiple models, should use first one
		chatter, err := New(config, "gpt-4", "gpt-4-turbo", "gpt-3.5-turbo")

		it.Then(t).
			Should(it.Nil(err)).
			ShouldNot(it.Nil(chatter))
	})
}

func TestNewErrorHandling(t *testing.T) {
	t.Run("invalid_schema", func(t *testing.T) {
		config := newMockConfig("invalid:provider/capability/family", "model")

		it.Then(t).Should(
			it.Error(New(config)).Contain("invalid schema"),
		)
	})

	t.Run("unsupported_configuration", func(t *testing.T) {
		config := newMockConfig("provider:unsupported/provider/family", "model")

		it.Then(t).Should(
			it.Error(New(config)).Contain("configuration is not supported"),
		)
	})

	t.Run("invalid_dimensions_bedrock_titan", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/embedding/titan", "titan-embed")
		config.Set(opt_dimensions, "invalid-number")
		config.Set(opt_region, "us-east-1")

		it.Then(t).Should(
			it.Error(New(config)).Contain("invalid config, dimensions"),
		)
	})

	t.Run("invalid_dimensions_openai_text2vec", func(t *testing.T) {
		config := newMockConfig("provider:openai/embedding/text2vec", "text-embedding-3-small")
		config.Set(opt_dimensions, "not-a-number")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		it.Then(t).Should(
			it.Error(New(config)).Contain("invalid config, dimensions"),
		)
	})

	t.Run("missing_dimensions_bedrock_titan", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/embedding/titan", "titan-embed")
		config.Set(opt_region, "us-east-1")
		// dimensions not set, should default to empty string and fail parsing

		it.Then(t).Should(
			it.Error(New(config)).Contain("invalid config, dimensions"),
		)
	})

	t.Run("missing_dimensions_openai_text2vec", func(t *testing.T) {
		config := newMockConfig("provider:openai/embedding/text2vec", "text-embedding-3-small")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")
		// dimensions not set

		it.Then(t).Should(
			it.Error(New(config)).Contain("invalid config, dimensions"),
		)
	})
}

func TestCurlFunction(t *testing.T) {
	t.Run("default_timeout", func(t *testing.T) {
		config := newMockConfig("provider:test/model", "test-model")

		client := curl(config)

		it.Then(t).Should(
			it.Equal(client.Timeout, 2*time.Minute),
		).
			ShouldNot(it.Nil(client))
	})

	t.Run("custom_timeout_seconds", func(t *testing.T) {
		config := newMockConfig("provider:test/model", "test-model")
		config.Set(opt_timeout, "30")

		client := curl(config)

		it.Then(t).Should(
			it.Equal(client.Timeout, 30*time.Second),
		).
			ShouldNot(it.Nil(client))
	})

	t.Run("invalid_timeout_uses_default", func(t *testing.T) {
		config := newMockConfig("provider:test/model", "test-model")
		config.Set(opt_timeout, "invalid-number")

		client := curl(config)

		it.Then(t).Should(
			it.Equal(client.Timeout, 2*time.Minute),
		).
			ShouldNot(it.Nil(client))
	})

	t.Run("empty_timeout_uses_default", func(t *testing.T) {
		config := newMockConfig("provider:test/model", "test-model")
		config.Set(opt_timeout, "")

		client := curl(config)

		it.Then(t).Should(
			it.Equal(client.Timeout, 2*time.Minute),
		).
			ShouldNot(it.Nil(client))
	})

	t.Run("zero_timeout", func(t *testing.T) {
		config := newMockConfig("provider:test/model", "test-model")
		config.Set(opt_timeout, "0")

		client := curl(config)

		it.Then(t).Should(
			it.Equal(client.Timeout, 0*time.Second),
		).
			ShouldNot(it.Nil(client))
	})
}

func TestConfigInterface(t *testing.T) {
	t.Run("mock_config_implementation", func(t *testing.T) {
		config := newMockConfig("provider:test/provider/capability/family", "model")
		config.Set("key1", "value1")
		config.Set("key2", "value2")

		// Test Config interface compliance
		var iface Config = config

		it.Then(t).Should(
			it.Equal(iface.Provider(), curie.IRI("provider:test/provider/capability/family")),
			it.Equal(iface.Model(), "model"),
			it.Equal(iface.Get("key1"), "value1"),
			it.Equal(iface.Get("key2"), "value2"),
			it.Equal(iface.Get("nonexistent"), ""),
		)
	})
}

func TestNetconfigImplementation(t *testing.T) {
	t.Run("netconfig_interface_compliance", func(t *testing.T) {
		// Create a mock netrc.Machine
		mockMachine := &mockNetrcMachine{
			data: map[string]string{
				opt_provider: "provider:openai/foundation/gpt",
				opt_model:    "gpt-4",
				"region":     "us-east-1",
				"host":       "https://api.openai.com",
				"secret":     "sk-test-key",
			},
		}

		config := &netconfigTest{machine: mockMachine}

		// Test Config interface compliance
		var iface Config = config

		it.Then(t).Should(
			it.Equal(iface.Provider(), curie.IRI("provider:openai/foundation/gpt")),
			it.Equal(iface.Model(), "gpt-4"),
			it.Equal(iface.Get("region"), "us-east-1"),
			it.Equal(iface.Get("host"), "https://api.openai.com"),
			it.Equal(iface.Get("secret"), "sk-test-key"),
			it.Equal(iface.Get("nonexistent"), ""),
		)
	})
}

func TestConfigurationParameterConstants(t *testing.T) {
	t.Run("verify_constants", func(t *testing.T) {
		it.Then(t).Should(
			it.Equal(opt_provider, "provider"),
			it.Equal(opt_model, "model"),
			it.Equal(opt_region, "region"),
			it.Equal(opt_host, "host"),
			it.Equal(opt_secret, "secret"),
			it.Equal(opt_dimensions, "dimensions"),
			it.Equal(opt_timeout, "timeout"),
		)
	})
}

func TestEdgeCasesAndBoundaryConditions(t *testing.T) {
	t.Run("empty_model_string", func(t *testing.T) {
		config := newMockConfig("", "")

		it.Then(t).Should(
			it.Error(New(config)).Contain("invalid schema"),
		)
	})

	t.Run("model_with_only_schema", func(t *testing.T) {
		config := newMockConfig("provider:", "")

		it.Then(t).Should(
			it.Error(New(config)).Contain("configuration is not supported"),
		)
	})

	t.Run("maximum_valid_dimensions", func(t *testing.T) {
		config := newMockConfig("provider:openai/embedding/text2vec", "text-embedding-3-large")
		config.Set(opt_dimensions, "3072") // Max for text-embedding-3-large
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		chatter, err := New(config)

		it.Then(t).Should(
			it.Nil(err),
		).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("zero_dimensions", func(t *testing.T) {
		config := newMockConfig("provider:openai/embedding/text2vec", "text-embedding-ada-002")
		config.Set(opt_dimensions, "0")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		chatter, err := New(config)

		it.Then(t).Should(
			it.Nil(err),
		).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("negative_dimensions", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/embedding/titan", "titan-embed-text-v1")
		config.Set(opt_dimensions, "-100")
		config.Set(opt_region, "us-east-1")

		chatter, err := New(config)

		it.Then(t).Should(
			it.Nil(err),
		).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("very_large_timeout", func(t *testing.T) {
		config := newMockConfig("provider:test/model", "test-model")
		config.Set(opt_timeout, "86400") // 24 hours in seconds

		client := curl(config)

		it.Then(t).Should(
			it.Equal(client.Timeout, 86400*time.Second),
		).
			ShouldNot(it.Nil(client))
	})
}

func TestComplexModelIdentifiers(t *testing.T) {
	t.Run("anthropic_claude_models", func(t *testing.T) {
		config := newMockConfig("provider:bedrock/foundation/converse", "anthropic.claude-3-5-sonnet-20241022-v2:0")
		config.Set(opt_region, "us-east-1")

		chatter, err := New(config)

		it.Then(t).Should(
			it.Nil(err),
		).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("versioned_model_identifier", func(t *testing.T) {
		config := newMockConfig("provider:openai/foundation/gpt", "gpt-4-0125-preview")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		chatter, err := New(config)

		it.Then(t).Should(
			it.Nil(err),
		).
			ShouldNot(it.Nil(chatter))
	})

	t.Run("custom_model_with_special_characters", func(t *testing.T) {
		config := newMockConfig("provider:openai/foundation/gpt", "custom-model_v1.2-beta")
		config.Set(opt_host, "https://api.openai.com")
		config.Set(opt_secret, "sk-test-key")

		chatter, err := New(config)

		it.Then(t).Should(
			it.Nil(err),
		).
			ShouldNot(it.Nil(chatter))
	})
}

// Mock implementation of netrc.Machine for testing
type mockNetrcMachine struct {
	data map[string]string
}

func (m *mockNetrcMachine) Get(key string) string {
	return m.data[key]
}

// Test implementation of netconfig for testing
type netconfigTest struct {
	machine *mockNetrcMachine
}

func (c *netconfigTest) Provider() curie.IRI {
	return curie.IRI(c.machine.Get(opt_provider))
}

func (c *netconfigTest) Model() string {
	return c.machine.Get(opt_model)
}

func (c *netconfigTest) Get(key string) string {
	return c.machine.Get(key)
}

func TestAllProviderTypes(t *testing.T) {
	providerTests := []struct {
		name     string
		provider string
		model    string
		setup    func(*mockConfig)
		hasErr   bool
	}{
		{
			name:     "bedrock_titan_embedding",
			provider: "provider:bedrock/embedding/titan",
			model:    "amazon.titan-embed-text-v1",
			setup: func(c *mockConfig) {
				c.Set(opt_dimensions, "1536")
				c.Set(opt_region, "us-east-1")
			},
			hasErr: false,
		},
		{
			name:     "bedrock_converse_anthropic",
			provider: "provider:bedrock/foundation/converse",
			model:    "anthropic.claude-3-sonnet-20240229-v1:0",
			setup: func(c *mockConfig) {
				c.Set(opt_region, "us-west-2")
			},
			hasErr: false,
		},
		{
			name:     "bedrock_llama",
			provider: "provider:bedrock/foundation/llama",
			model:    "meta.llama3-8b-instruct-v1:0",
			setup: func(c *mockConfig) {
				c.Set(opt_region, "eu-central-1")
			},
			hasErr: false,
		},
		{
			name:     "bedrock_nova_pro",
			provider: "provider:bedrock/foundation/nova",
			model:    "amazon.nova-pro-v1:0",
			setup: func(c *mockConfig) {
				c.Set(opt_region, "us-east-1")
			},
			hasErr: false,
		},
		{
			name:     "openai_text_embedding_ada",
			provider: "provider:openai/embedding/text2vec",
			model:    "text-embedding-ada-002",
			setup: func(c *mockConfig) {
				c.Set(opt_dimensions, "1536")
				c.Set(opt_host, "https://api.openai.com")
				c.Set(opt_secret, "sk-proj-test123")
			},
			hasErr: false,
		},
		{
			name:     "openai_gpt4_omni",
			provider: "provider:openai/foundation/gpt",
			model:    "gpt-4o",
			setup: func(c *mockConfig) {
				c.Set(opt_host, "https://api.openai.com")
				c.Set(opt_secret, "sk-proj-test456")
			},
			hasErr: false,
		},
		{
			name:     "unsupported_provider",
			provider: "provider:mistral/foundation/mixtral",
			model:    "mixtral-8x7b",
			setup:    func(c *mockConfig) {},
			hasErr:   true,
		},
	}

	for _, tc := range providerTests {
		t.Run(tc.name, func(t *testing.T) {
			config := newMockConfig(tc.provider, tc.model)
			tc.setup(config)

			chatter, err := New(config)

			if tc.hasErr {
				it.Then(t).Should(
					it.Nil(chatter),
				).
					ShouldNot(it.Nil(err))
			} else {
				it.Then(t).Should(
					it.Nil(err),
				).
					ShouldNot(it.Nil(chatter))
			}
		})
	}
}
