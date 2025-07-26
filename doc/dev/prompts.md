# Prompts for Copilot

Prompts below guides coding agents with the library development and maintanace. 


## Implement or refactor the model provider

Your task is to **implement an adapter for the Amazon Nova LLM via the AWS Bedrock Invoke API**.

The adapter serves a single purpose: it consumes the specific API of the Nova model and exposes it through the **common API defined by this library**.
You must implement the adapter **using the primitives defined in the `aio/provider` submodule**.

#### To complete this task, you need to:

1. **Understand the common API** defined at the root of this library.
   Key concepts to study: `Content`, `Message`, `Prompt`, `Reply`, etc.
2. **Understand the APIs required for provider (adapter) development.**
3. **Study the specifics of the Nova API**:
   [Nova Invoke API Documentation](https://docs.aws.amazon.com/nova/latest/userguide/using-invoke-api.html)
4. **Use the implementation of `provider/bedrock/llm/llama` as a reference** for your work.
   Adhere **strictly** to this example — your implementation should follow the same patterns and structure.

#### Implementation Requirements:

* Place your implementation in:
  `provider/bedrock/llm/nova`
* Implement and expose the main abstractions:
  `Factory`, `Encoder`, `Decoder`, etc.
* Follow the reference example (`llama`) as closely as possible — **do not deviate from its implementation**.
  Make your code as similar (and identical where possible) to the example, adjusted only to fit the Nova API.



## Implement unit testing for encoder

Your task is to **implement unit tests for `encoder.go`**.
The tests should be **compact and readable for humans**.

#### Strict requirements:

1. You **must** use the assertion library:
   `github.com/fogfish/it/v2`

2. Use the **JSON matcher** from the same library to ensure compatibility with the model protocol.
   The JSON matcher is a simple and effective way to verify the *shape* of the expected object.
   It accepts any Go object as actual value and compares it to the expected JSON structure, provided as a valid JSON string.
   You can learn more about its syntax in the *“JSON matchers”* section of the [README](https://raw.githubusercontent.com/fogfish/it/refs/heads/main/README.md).

3. Do **not** use wildcard matchers for values in the JSON object — especially for prompt messages.
   Instead, use **regular expression matches within JSON strings**.
   To match a string field against a regular expression, use the following pattern:

   ```json
   { "field": "regex:your-valid-golang-regex" }
   ```

   (The `regex:` prefix indicates to the assertion library that the value is a regular expression.)

4. Do **not** write a dedicated unit test for each function in the interface.
   Instead, group related functionality into **logical blocks**.


#### Testing Pattern:

Follow this pattern when writing your tests:

```go
f, err := factory()
it.Then(t).Must(it.Nil(err))

f.WithXxx(...)
f.AsXxx(...)

it.Then(t).Should(
  it.Json(f.Build()).Equiv(/* json pattern here */)
)
```

Make sure the tests are clear, minimal, and maintainable, adhering strictly to the above style and constraints.


## Implement unit testing for decoder

Your task is to **implement unit tests for `decoder.go`**.
The tests should be **compact and readable for humans**.

#### Strict requirements:

1. You **must** use the assertion library:
   `github.com/fogfish/it/v2`

2. Use the **JSON matcher** from the same library to validate content of chatter.Reply instead of implementing connection of native assertions
   The JSON matcher is a simple and effective way to verify the *shape* of the expected object.
   It accepts any Go object as actual value and compares it to the expected JSON structure, provided as a valid JSON string.
   You can learn more about its syntax in the *“JSON matchers”* section of the [README](https://raw.githubusercontent.com/fogfish/it/refs/heads/main/README.md).

3. Do **not** use wildcard matchers for values in the JSON object — especially for prompt messages.
   Instead, use **regular expression matches within JSON strings**.
   To match a string field against a regular expression, use the following pattern:

   ```json
   { "field": "regex:your-valid-golang-regex" }
   ```

   (The `regex:` prefix indicates to the assertion library that the value is a regular expression.)

4. Do **not** write a dedicated unit test for each function in the interface.
   Instead, group related functionality into **logical blocks**.

#### Testing Pattern:

Follow this pattern when writing your tests:

```go
input := &reply{/* complete structure  */}
reply, err := decoder{}.Decode(input)


it.Then(t).Should(
  it.Nil(err),
  it.Json(f.Build()).Equiv(/* json pattern here */)
)
```

Make sure the tests are clear, minimal, and maintainable, adhering strictly to the above style and constraints.
