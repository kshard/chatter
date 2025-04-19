//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrock

import (
	"fmt"
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsbedrock"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type InferenceProfile struct {
	constructs.Construct
	profile *string
	llm     awsbedrock.FoundationModel
}

func NewInferenceProfile(scope constructs.Construct, id *string, profile *string) *InferenceProfile {
	if profile == nil {
		panic(fmt.Errorf("undefined inference profile"))
	}

	seq := strings.SplitN(*profile, ".", 2)
	if len(seq) != 2 {
		panic(fmt.Errorf("invalid inference profile"))
	}

	c := &InferenceProfile{Construct: constructs.NewConstruct(scope, id)}
	c.profile = profile
	c.llm = awsbedrock.FoundationModel_FromFoundationModelId(
		c.Construct,
		jsii.String("LLM"),
		awsbedrock.NewFoundationModelIdentifier(jsii.String(seq[1])),
	)

	return c
}

func (c *InferenceProfile) GrantAccessIn(grantee awsiam.IGrantable, region *string) {
	parn := awscdk.Stack_Of(c.Construct).FormatArn(
		&awscdk.ArnComponents{
			ArnFormat:    awscdk.ArnFormat_SLASH_RESOURCE_NAME,
			Service:      jsii.String("bedrock"),
			Account:      awscdk.Aws_ACCOUNT_ID(),
			Region:       region,
			Resource:     jsii.String("inference-profile"),
			ResourceName: c.profile,
		},
	)

	larn := awscdk.Stack_Of(c.Construct).FormatArn(
		&awscdk.ArnComponents{
			ArnFormat:    awscdk.ArnFormat_SLASH_RESOURCE_NAME,
			Service:      jsii.String("bedrock"),
			Account:      jsii.String(""),
			Region:       jsii.String("*"),
			Resource:     jsii.String("foundation-model"),
			ResourceName: c.llm.ModelId(),
		},
	)

	awsiam.Grant_AddToPrincipal(
		&awsiam.GrantOnPrincipalOptions{
			Grantee:      grantee,
			Actions:      jsii.Strings("bedrock:InvokeModel"),
			ResourceArns: jsii.Strings(*parn, *larn),
		},
	)
}

//------------------------------------------------------------------------------

type FoundationModel struct {
	constructs.Construct
	llm awsbedrock.FoundationModel
}

func NewFoundationModel(scope constructs.Construct, id *string, foundationModelId awsbedrock.FoundationModelIdentifier) *FoundationModel {
	c := &FoundationModel{Construct: constructs.NewConstruct(scope, id)}
	c.llm = awsbedrock.FoundationModel_FromFoundationModelId(
		c.Construct,
		jsii.String("LLM"),
		foundationModelId,
	)

	return c
}

func (c *FoundationModel) GrantAccess(grantee awsiam.IGrantable) {
	awsiam.Grant_AddToPrincipal(
		&awsiam.GrantOnPrincipalOptions{
			Grantee:      grantee,
			Actions:      jsii.Strings("bedrock:InvokeModel"),
			ResourceArns: jsii.Strings(*c.llm.ModelArn()),
		},
	)
}

func (c *FoundationModel) GrantAccessIn(grantee awsiam.IGrantable, region *string) {
	arn := awscdk.Stack_Of(c.Construct).FormatArn(
		&awscdk.ArnComponents{
			ArnFormat:    awscdk.ArnFormat_SLASH_RESOURCE_NAME,
			Service:      jsii.String("bedrock"),
			Account:      jsii.String(""),
			Region:       region,
			Resource:     jsii.String("foundation-model"),
			ResourceName: c.llm.ModelId(),
		},
	)

	awsiam.Grant_AddToPrincipal(
		&awsiam.GrantOnPrincipalOptions{
			Grantee:      grantee,
			Actions:      jsii.Strings("bedrock:InvokeModel"),
			ResourceArns: jsii.Strings(*arn),
		},
	)
}
