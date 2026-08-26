//
// Copyright (C) 2024 - 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package iam

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type BatchRoleProps struct {
	BucketName string

	// InferenceProfile grants permission to invoke a cross-region (or
	// application) inference profile via batch inference. Required when
	// the batch job's model ID is an inference-profile ID rather than a
	// bare foundation model ID (e.g. "us.anthropic.claude-haiku-4-5-..."
	// instead of "anthropic.claude-haiku-4-5-..."), since some models
	// only support batch inference through such a profile. Bedrock then
	// needs permission both on the profile and on the foundation model it
	// routes to, in whichever region that turns out to be.
	// https://docs.aws.amazon.com/bedrock/latest/userguide/batch-iam-sr.html#batch-iam-sr-ip
	InferenceProfile *InferenceProfileAccess
}

// InferenceProfileAccess identifies a cross-region inference profile and
// the underlying foundation model it routes to, for [BatchRoleProps].
type InferenceProfileAccess struct {
	// ID of the inference profile, e.g.
	// "us.anthropic.claude-haiku-4-5-20251001-v1:0".
	ProfileID string

	// ID of the underlying foundation model, e.g.
	// "anthropic.claude-haiku-4-5-20251001-v1:0".
	ModelID string
}

func NewBatchRole(scope constructs.Construct, id *string, props *BatchRoleProps) awsiam.Role {
	role := awsiam.NewRole(scope, id,
		&awsiam.RoleProps{
			AssumedBy: awsiam.NewServicePrincipal(
				jsii.String("bedrock.amazonaws.com"),
				&awsiam.ServicePrincipalOpts{
					Conditions: &map[string]any{
						"StringEquals": map[string]*string{
							"aws:SourceAccount": awscdk.Aws_ACCOUNT_ID(),
						},
						"ArnEquals": map[string]*string{
							"aws:SourceArn": awscdk.Stack_Of(scope).FormatArn(
								&awscdk.ArnComponents{
									Service:      jsii.String("bedrock"),
									Region:       awscdk.Aws_REGION(),
									Account:      awscdk.Aws_ACCOUNT_ID(),
									Resource:     jsii.String("model-invocation-job"),
									ResourceName: jsii.String("*"),
								},
							),
						},
					},
				},
			),
		},
	)

	role.AddToPolicy(
		awsiam.NewPolicyStatement(
			&awsiam.PolicyStatementProps{
				Actions: jsii.Strings("s3:GetObject", "s3:PutObject", "s3:ListBucket"),
				Resources: &[]*string{
					awscdk.Stack_Of(scope).FormatArn(
						&awscdk.ArnComponents{
							Service:  jsii.String("s3"),
							Region:   jsii.String(""),
							Account:  jsii.String(""),
							Resource: jsii.String(props.BucketName),
						},
					),
					awscdk.Stack_Of(scope).FormatArn(
						&awscdk.ArnComponents{
							Service:      jsii.String("s3"),
							Region:       jsii.String(""),
							Account:      jsii.String(""),
							Resource:     jsii.String(props.BucketName),
							ResourceName: jsii.String("*"),
						},
					),
				},
			},
		),
	)

	if props.InferenceProfile != nil {
		role.AddToPolicy(
			awsiam.NewPolicyStatement(
				&awsiam.PolicyStatementProps{
					Actions: jsii.Strings("bedrock:InvokeModel"),
					Resources: &[]*string{
						awscdk.Stack_Of(scope).FormatArn(
							&awscdk.ArnComponents{
								Service:      jsii.String("bedrock"),
								Resource:     jsii.String("inference-profile"),
								ResourceName: jsii.String(props.InferenceProfile.ProfileID),
							},
						),
						awscdk.Stack_Of(scope).FormatArn(
							&awscdk.ArnComponents{
								Service:      jsii.String("bedrock"),
								Region:       jsii.String("*"),
								Account:      jsii.String(""),
								Resource:     jsii.String("foundation-model"),
								ResourceName: jsii.String(props.InferenceProfile.ModelID),
							},
						),
					},
				},
			),
		)
	}

	return role
}
