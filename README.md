<p align="center">
  <h3 align="center">chatter</h3>
  <p align="center"><strong>adapter over LLMs interface</strong></p>

  <p align="center">
    <!-- Build Status  -->
    <a href="https://github.com/kshard/chatter/actions/">
      <img src="https://github.com/kshard/chatter/workflows/build/badge.svg" />
    </a>
    <!-- GitHub -->
    <a href="https://github.com/kshard/chatter">
      <img src="https://img.shields.io/github/last-commit/kshard/chatter.svg" />
    </a>
    <!-- Coverage -->
    <a href="https://coveralls.io/github/kshard/chatter?branch=main">
      <img src="https://coveralls.io/repos/github/kshard/chatter/badge.svg?branch=main" />
    </a>
    <!-- Go Card -->
    <a href="https://goreportcard.com/report/github.com/kshard/chatter">
      <img src="https://goreportcard.com/badge/github.com/kshard/chatter" />
    </a>
  </p>

  <table align="center">
    <thead><tr><th>sub-module</th><th>doc</th><th>about</th></tr></thead>
    <tbody>
    <!-- Module chatter types -->
    <tr><td><a href=".">
      <img src="https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=v*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/kshard/chatter">
      <img src="https://img.shields.io/badge/doc-chatter-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      chatter types
    </td></tr>
    <!-- Module bedrock -->
    <tr><td><a href=".">
      <img src="https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=llm/bedrock/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/kshard/chatter/llm/bedrock">
      <img src="https://img.shields.io/badge/doc-bedrock-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      AWS Bedrock LLMs
    </td></tr>
    <!-- Module bedrock batch -->
    <tr><td><a href=".">
      <img src="https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=llm/bedrockbatch/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/kshard/chatter/llm/bedrockbatch">
      <img src="https://img.shields.io/badge/doc-bedrockbatch-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      AWS Bedrock Batch Inference
    </td></tr>
		<!-- Module openai -->
    <tr><td><a href=".">
      <img src="https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=llm/openai/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/kshard/chatter/llm/openai">
      <img src="https://img.shields.io/badge/doc-openai-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      OpenAI and compatible LLMs
    </td></tr>
		</tbody>
	</table>
</p>

---

`chatter` is an adapter library that integrates with popular Large Language Models (LLMs) and hosting solutions optimized for text generation. It supports AWS Bedrock, OpenAI, and OpenAI-compatible models, providing a unified interface for seamless interaction.

In addition to model integration, `chatter` uses composable utilities such as caching, rate-limiting, and more, enabling efficient and scalable AI-powered applications.


## Inspiration

> A good prompt has 4 key elements: Role, Task, Requirements, Instructions.
["Are You AI Ready? Investigating AI Tools in Higher Education – Student Guide"](https://ucddublin.pressbooks.pub/StudentResourcev1_od/chapter/the-structure-of-a-good-prompt/)

In the research community, there was an attempt for making [standardized taxonomy of prompts](https://aclanthology.org/2023.findings-emnlp.946.pdf) for large language models (LLMs) to solve complex tasks. It encourages the community to adopt the TELeR taxonomy to achieve meaningful comparisons among LLMs, facilitating more accurate conclusions and helping the community achieve consensus on state-of-the-art LLM performance more efficiently.

The library addresses the LLMs comparisons by 
* Creating generic trait to "interact" with LLMs;
* Enabling prompt definition into [seven distinct levels](https://aclanthology.org/2023.findings-emnlp.946.pdf);
* Supporting variety of LLMs.   

```go
type Chatter interface {
	Prompt(context.Context, []fmt.Stringer, ...func(*Options)) (string, error)
}
```

## Getting started

The latest version of the library is available at `main` branch of this repository. All development, including new features and bug fixes, take place on the `main` branch using forking and pull requests as described in contribution guidelines. The stable version is available via Golang modules.

```go
package main

import (
	"context"
	"fmt"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/llm/bedrock"
)

func main() {
	assistant, err := bedrock.New(
		bedrock.WithLLM(bedrock.LLAMA3_0_8B_INSTRUCT),
	)
	if err != nil {
		panic(err)
	}

	var prompt chatter.Prompt
	prompt.WithTask("Extract keywords from the text: %s", /* ... */)

	reply, err := assistant.Prompt(context.Background(), prompt.ToSeq())
	if err != nil {
		panic(err)
	}

	fmt.Printf("==> (%d)\n%s\n", assistant.ConsumedTokens(), reply)
}
```

**Using LM Studio** `openai` module is also enables consumption of OpenAI compatible interfaces (e.g. LM Studio). It only requires explicit configuration of the host where model is hosted.

```go
import (
	"github.com/kshard/chatter/llm/openai"
)

assistant, err := openai.New(
  openai.WithHost("http://localhost:1234"),
  openai.WithLLM(openai.LLM("gemma-3-27b-it")),
)
```

**Using AWS Bedrock Inference Profiles** See the [explanation about usage of models with inference profile](https://repost.aws/questions/QUEU82wbYVQk2oU4eNwyiong/bedrock-api-invocation-error-on-demand-throughput-isn-s-supported)

```go
bedrock.New(
	bedrock.WithLLM("us." + bedrock.LLAMA3_0_8B_INSTRUCT),
)
```

## Prompt

Package `chatter` provides utilities for creating and managing structured prompts for language models.

The Prompt type allows you to create a structured prompt with various sections such as **task**, **rules**, **feedback**, **examples**, **context**, and **input**. This helps in maintaining semi-structured prompts while enabling efficient serialization into textual prompts.

```
{task}. {guidelines}.
1. {requirements}
2. {requirements}
3. ...
{feedback}
{examples}
{context}
{context}
...
{input}
```

Example usage:

```go
var prompt chatter.Prompt{}

prompt.WithTask("Translate the following text")

// Creates a guide section with the given note and text.
// It is complementary paragraph to the task.
prompt.WithGuide("Please translate the text accurately")

// Creates a rules / requirements section with the given note and text.
prompt.WithRules(
  "Strictly adhere to the following requirements when generating a response.",
  "Do not use any slang or informal language",
  "Do not invent new, unkown words",
)

// Creates a feedback section with the given note and text.
prompt.WithFeedback(
  "Improve the response based on feedback",
  "Previous translations were too literal.",
)

// Create example of input and expected output.
prompt.WithExample(`["Hello"]`, `["Hola"]`)

// Creates a context section with the given note and text.
prompt.WithContext(
  "Below are additional context relevant to your goal task.",
  "The text is a formal letter",
)

// Creates an input section with the given note and text.
prompt.WithInput(
  "Translate the following sentence",
  "Hello, how are you?",
)
```


## How To Contribute

The library is [MIT](LICENSE) licensed and accepts contributions via GitHub pull requests:

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Added some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create new Pull Request

The build and testing process requires [Go](https://golang.org) version 1.21 or later.

**build** and **test** library.

```bash
git clone https://github.com/kshard/chatter
cd chatter
go test ./...
```

### commit message

The commit message helps us to write a good release note, speed-up review process. The message should address two question what changed and why. The project follows the template defined by chapter [Contributing to a Project](http://git-scm.com/book/ch5-2.html) of Git book.

### bugs

If you experience any issues with the library, please let us know via [GitHub issues](https://github.com/kshard/chatter/issue). We appreciate detailed and accurate reports that help us to identity and replicate the issue. 


## License

[![See LICENSE](https://img.shields.io/github/license/kshard/chatter.svg?style=for-the-badge)](LICENSE)


## Reference
1. [Use a tool to complete an Amazon Bedrock model response](https://docs.aws.amazon.com/bedrock/latest/userguide/tool-use.html)
2. [Converse API tool use examples](https://docs.aws.amazon.com/bedrock/latest/userguide/tool-use-examples.html)
3. [Call a tool with the Converse API](https://docs.aws.amazon.com/bedrock/latest/userguide/tool-use-inference-call.html)
  