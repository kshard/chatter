# chatter

The library is adapter over various popular Large Language Models (LLMs) tuned for text generation: AWS BedRock, OpenAI.

[![Version](https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=v*)](https://github.com/kshard/chatter/releases)
[![Documentation](https://pkg.go.dev/badge/github.com/kshard/chatter)](https://pkg.go.dev/github.com/kshard/chatter)
[![Build Status](https://github.com/kshard/chatter/workflows/build/badge.svg)](https://github.com/kshard/chatter/actions/)
[![Git Hub](https://img.shields.io/github/last-commit/kshard/chatter.svg)](https://github.com/kshard/chatter)
[![Coverage Status](https://coveralls.io/repos/github/kshard/chatter/badge.svg?branch=main)](https://coveralls.io/github/kshard/chatter?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/kshard/chatter)](https://goreportcard.com/report/github.com/kshard/chatter)

| module                                          | version                                                                                                                                     | api                                                                                                                          |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| [github.com/kshard/chatter/openai](./openai/)   | [![Version](https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=openai/*)](https://github.com/kshard/chatter/releases)  | [![Documentation](https://pkg.go.dev/badge/github.com/kshard/chatter)](https://pkg.go.dev/github.com/kshard/chatter/openai)  |
| [github.com/kshard/chatter/bedrock](./bedrock/) | [![Version](https://img.shields.io/github/v/tag/kshard/chatter?label=version&filter=bedrock/*)](https://github.com/kshard/chatter/releases) | [![Documentation](https://pkg.go.dev/badge/github.com/kshard/chatter)](https://pkg.go.dev/github.com/kshard/chatter/bedrock) |


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
	Prompt(context.Context, *Prompt) (*Prompt, error)
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
		bedrock.WithModel(bedrock.LLAMA3_0_8B_INSTRUCT),
		bedrock.WithQuotaTokensInReply(512),
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



<!--



It also defines a common data type for prompting, which is automatically adopted into the most suitable prompt style of the target model.

```go
prompt := chatter.NewPrompt(
  chatter.WithStratum("You are pirate, Captain Blood."),
  chatter.WithContext("\"Captain Blood: His Odyssey\" book by Rafael Sabatini constraints replies."),
)

prompt.Inquiry("What we are doing upon arrival on the island of Barbados?")
```

The prompt above settle the model behavior and context into [Captain Blood Novel](https://en.wikipedia.org/wiki/Captain_Blood_(novel))

**gpt-3.5-turbo-0125**: Ah, me hearties! Upon reaching the shores of Barbados, we shall make haste to investigate the town of Bridgetown. Keep a keen eye out for opportunities to bolster our crew and provisions, for Barbados is a veritable treasure trove of resources ripe for the taking. And beware the cunning ways of the local authorities, for they may seek to thwart our plans at every turn. But fear not, for with cunning and daring, we shall carve a path to greatness on this island paradise. Onward, me hearties, for our odyssey has only just begun!

**llama2-13b-chat-v1**: Ahoy, matey! *adjusts eye patch* Ah, the fair isle of Barbados! *smirks wickedly* We be here for plunder, of course! The Spanish galleons be filled with gold and treasure, just waitin' for us to take 'em down. But first, we'll need to gather our wits and our strength, and prepare for battle. *pours self a mug of grog* Care to join me in a drink, matey? We'll need all the courage we can get for what's to come!

**titan-text-lite-v1**: We are looking for a ship to command.



-->