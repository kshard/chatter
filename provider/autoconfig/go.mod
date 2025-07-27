module github.com/kshard/chatter/provider/autoconfig

go 1.23.1

toolchain go1.24.5

replace github.com/kshard/chatter => ../../

replace github.com/kshard/chatter/provider/bedrock => ../bedrock

replace github.com/kshard/chatter/provider/openai => ../openai

require (
	github.com/fogfish/curie/v2 v2.1.2
	github.com/fogfish/gurl/v2 v2.10.0
	github.com/fogfish/it/v2 v2.2.4
	github.com/jdxcode/netrc v1.0.0
	github.com/kshard/chatter v0.0.0-00010101000000-000000000000
	github.com/kshard/chatter/provider/bedrock v0.0.0-00010101000000-000000000000
	github.com/kshard/chatter/provider/openai v0.0.0-00010101000000-000000000000
)

require (
	github.com/ajg/form v1.5.2-0.20200323032839-9aeb3cf462e1 // indirect
	github.com/aws/aws-sdk-go-v2 v1.36.6 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.11 // indirect
	github.com/aws/aws-sdk-go-v2/config v1.29.18 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.71 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.33 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.3.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.6.37 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/bedrockruntime v1.31.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.12.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.12.18 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.6 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.4 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.34.1 // indirect
	github.com/aws/smithy-go v1.22.4 // indirect
	github.com/fogfish/faults v0.3.2 // indirect
	github.com/fogfish/golem/hseq v1.3.0 // indirect
	github.com/fogfish/golem/optics v0.14.0 // indirect
	github.com/fogfish/logger/v3 v3.2.0 // indirect
	github.com/fogfish/logger/x/xlog v0.0.1 // indirect
	github.com/fogfish/opts v0.0.5 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	golang.org/x/net v0.17.0 // indirect
)
