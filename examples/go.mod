module github.com/kshard/chatter/examples

go 1.21.0

replace github.com/kshard/chatter => ../

replace github.com/kshard/chatter/openai => ../openai

replace github.com/kshard/chatter/bedrock => ../bedrock

require (
	github.com/kshard/chatter v0.0.4
	github.com/kshard/chatter/bedrock v0.0.0-00010101000000-000000000000
)

require (
	github.com/aws/aws-sdk-go-v2 v1.25.2 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.1 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.27.5 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.5 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.15.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.11.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.11.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.20.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.23.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.28.2 // indirect
	github.com/aws/smithy-go v1.20.1 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
)
