# HowTo: autoconfig

`autoconfig` is a configuration loader for LLM instances. It reads named provider
specifications from a file (YAML, JSON, or netrc) and builds ready-to-use
`chatter.Chatter` values that your application retrieves by name.

## Installation

```
go get github.com/kshard/chatter/provider/autoconfig
```

## Configuration file formats

Every format maps a **logical name** to an **instance specification**.  
A specification always requires two fields: `provider` and `model`.

### YAML — `~/.llm.yaml`

```yaml
# chat model on AWS Bedrock
claude:
  provider: "provider:bedrock/foundation/converse"
  model: "anthropic.claude-3-5-sonnet-20241022-v2:0"
  region: "us-east-1"

# chat model on OpenAI
gpt4:
  provider: "provider:openai/foundation/gpt"
  model: "gpt-4o"
  host: "https://api.openai.com"
  secret: "sk-..."

# embedding model on AWS Bedrock
embed:
  provider: "provider:bedrock/embedding/titan"
  model: "amazon.titan-embed-text-v2:0"
  region: "us-east-1"
  dimensions: 1024
```

### JSON — `~/.llm.json`

```json
{
  "claude": {
    "provider": "provider:bedrock/foundation/converse",
    "model": "anthropic.claude-3-5-sonnet-20241022-v2:0",
    "region": "us-east-1"
  },
  "gpt4": {
    "provider": "provider:openai/foundation/gpt",
    "model": "gpt-4o",
    "host": "https://api.openai.com",
    "secret": "sk-..."
  }
}
```

### netrc — `~/.llm` (default for any other extension)

```
machine claude
  provider provider:bedrock/foundation/converse
  model    anthropic.claude-3-5-sonnet-20241022-v2:0
  region   us-east-1

machine gpt4
  provider provider:openai/foundation/gpt
  model    gpt-4o
  host     https://api.openai.com
  secret   sk-...
```

## Supported providers

| `provider` value                       | Notes                                                                     |
| -------------------------------------- | ------------------------------------------------------------------------- |
| `provider:bedrock/foundation/converse` | AWS Bedrock Converse API — Anthropic Claude, Amazon Nova, etc.            |
| `provider:bedrock/foundation/llama`    | AWS Bedrock — Meta Llama models                                           |
| `provider:bedrock/foundation/nova`     | AWS Bedrock — Amazon Nova models                                          |
| `provider:bedrock/embedding/titan`     | AWS Bedrock — Amazon Titan embeddings; requires `dimensions`              |
| `provider:openai/foundation/gpt`       | OpenAI-compatible chat; requires `host` and `secret`                      |
| `provider:openai/embedding/text2vec`   | OpenAI-compatible embeddings; requires `host`, `secret`, and `dimensions` |
| `provider:google/foundation/gemini`    | Google Gemini; requires `secret`                                          |
| `provider:google/foundation/imagen`    | Google Imagen; requires `secret`                                          |

Optional fields shared by all providers:

| Field        | Default         | Meaning                                      |
| ------------ | --------------- | -------------------------------------------- |
| `region`     | AWS SDK default | AWS region (Bedrock only)                    |
| `host`       | —               | Base API URL (OpenAI-compatible only)        |
| `secret`     | —               | API key                                      |
| `timeout`    | 120             | HTTP timeout in seconds                      |
| `dimensions` | —               | Embedding dimensions (embedding models only) |

## Loading instances in Go

### From a file path on the local filesystem

```go
import "github.com/kshard/chatter/provider/autoconfig"

// Returns (*Instances, error)
cfg, err := autoconfig.FromConfig("~/.llm.yaml")

// Panic variant — convenient in main()
cfg := autoconfig.MustFromConfig("~/.llm.yaml")
```

`FromConfig` / `MustFromConfig` resolve `~` to the home directory and detect the
format from the file extension (`.yaml` / `.yml` → YAML, `.json` → JSON,
anything else → netrc).

### From an `fs.FS` (e.g. in tests)

```go
import (
    "testing/fstest"
    "github.com/kshard/chatter/provider/autoconfig"
)

fsys := fstest.MapFS{
    "config.yaml": {Data: []byte(yamlContent)},
}
cfg, err := autoconfig.FromFile(fsys, "config.yaml")
```

### From an `io.Reader`

```go
f, _ := os.Open("config.json")
cfg, err := autoconfig.FromJSON(f)
```

Equivalent `Must*` variants exist for every loader (`MustFromFile`,
`MustFromYAML`, `MustFromJSON`).

## Using instances

```go
// Look up a named LLM
llm, ok := cfg.Model("claude")
if !ok {
    log.Fatal("instance 'claude' not found in config")
}

// Use it like any chatter.Chatter
reply, err := llm.Prompt(ctx, []chatter.Message{
    chatter.Stratum("You are a helpful assistant."),
    &chatter.Prompt{Task: "Translate 'hello' to French."},
})

// Aggregate token usage across all instances
total, perInstance := cfg.Usage()
log.Printf("total input tokens: %d", total.InputTokens)
```

## Testing with the mock provider

Use `provider:mock` in your config files to get an echo LLM that requires no
credentials and accumulates token usage — useful for unit tests.

```yaml
# test-config.yaml
bot:
  provider: "provider:mock"
  model: "echo"
```

```go
fsys := fstest.MapFS{
    "test-config.yaml": {Data: []byte(content)},
}
cfg := autoconfig.MustFromFile(fsys, "test-config.yaml")
llm, _ := cfg.Model("bot")
reply, _ := llm.Prompt(ctx, []chatter.Message{chatter.Stratum("ping")})
// reply.String() == "ping"
```
