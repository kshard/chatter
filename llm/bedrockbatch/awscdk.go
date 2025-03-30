//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package bedrockbatch

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type RoleProps struct {
	BucketName string
}

func NewInferenceRole(scope constructs.Construct, id *string, props *RoleProps) awsiam.Role {
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
							Resource: jsii.String(props.BucketName),
						},
					),
					awscdk.Stack_Of(scope).FormatArn(
						&awscdk.ArnComponents{
							Service:      jsii.String("s3"),
							Resource:     jsii.String(props.BucketName),
							ResourceName: jsii.String("*"),
						},
					),
				},
			},
		),
	)

	return role
}
