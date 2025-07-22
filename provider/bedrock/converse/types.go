package converse

import "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"

type encoder struct {
	req *bedrockruntime.ConverseInput
}

type decoder struct{}

// func New(model string)
