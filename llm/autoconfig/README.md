# LLM autoconfig

Uses `~/.netrc` to configure LLM API.
```go
autoconfig.New("myservice")
```

Configure AWS Bedrock API
```
myservice
  provider bedrock
  region us-east-1
  family llama3
  model meta.llama3-1-70b-instruct-v1:0
```

Configure AWs Bedrock Converse API
```
myservice
  provider converse
  region us-east-1
  model us.anthropic.claude-3-7-sonnet-20250219-v1:0
```

Configure OpenAI API
```
myservice
  provider openai
  host https://api.openai.com
  secret sk-...AIA
  model gpt-4o
```

Configure LM Studio
```
myservice
  provider openai
  host http://localhost:1234
  model gemma-3-24b-it
```