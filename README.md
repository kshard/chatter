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
      <img src="https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=bedrock/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/kshard/chatter/bedrock">
      <img src="https://img.shields.io/badge/doc-bedrock-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      AWS Bedrock LLMs
    </td></tr>
    <!-- Module bedrock batch -->
    <tr><td><a href=".">
      <img src="https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=bedrockbatch/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/kshard/chatter/bedrockbatch">
      <img src="https://img.shields.io/badge/doc-bedrockbatch-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      AWS Bedrock Batch Inference
    </td></tr>
		<!-- Module openai -->
    <tr><td><a href=".">
      <img src="https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=openai/*"/>
    </a></td>
    <td><a href="https://pkg.go.dev/github.com/kshard/chatter/openai">
      <img src="https://img.shields.io/badge/doc-openai-007d9c?logo=go&logoColor=white&style=platic" />
    </a></td>
    <td>
      OpenAI LLMs
    </td></tr>
		</tbody>
	</table>
</p>

---


The library is adapter over various popular Large Language Models (LLMs) tuned for text generation: AWS BedRock, OpenAI.


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
	Prompt(context.Context, encoding.TextMarshaler, ...func(*Options)) (string, error)
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
	"github.com/kshard/chatter/bedrock"
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

	reply, err := assistant.Prompt(context.Background(), &prompt)
	if err != nil {
		panic(err)
	}

	fmt.Printf("==> (%d)\n%s\n", assistant.ConsumedTokens(), reply)
}
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

