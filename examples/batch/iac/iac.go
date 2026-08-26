//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

// CDK application provisioning the AWS resources required to run the
// ../batch example: an S3 bucket to stage batch input/output JSONL files,
// and the IAM role Bedrock assumes to read/write that bucket while running
// a batch inference job (see provider/bedrock/iam.NewBatchRole).
//
//	cdk deploy
//
// After deployment, use the stack's BucketName and BatchRoleArn outputs as
// arguments to the batch example.
package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/kshard/chatter/provider/bedrock/iam"
)

type IacStackProps struct {
	awscdk.StackProps
}

func NewIacStack(scope constructs.Construct, id string, props *IacStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	bucket := awss3.NewBucket(stack, jsii.String("Bucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
	})

	// Keep in sync with the model constant in ../batch.go: it's a
	// cross-region inference profile, so the role needs bedrock:InvokeModel
	// on both the profile and the underlying foundation model.
	role := iam.NewBatchRole(stack, jsii.String("BatchRole"), &iam.BatchRoleProps{
		BucketName: *bucket.BucketName(),
		InferenceProfile: &iam.InferenceProfileAccess{
			ProfileID: "us.anthropic.claude-haiku-4-5-20251001-v1:0",
			ModelID:   "anthropic.claude-haiku-4-5-20251001-v1:0",
		},
	})

	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value: bucket.BucketName(),
	})
	awscdk.NewCfnOutput(stack, jsii.String("BatchRoleArn"), &awscdk.CfnOutputProps{
		Value: role.RoleArn(),
	})

	return stack
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	NewIacStack(app, "ChatterBatchExample", &IacStackProps{
		StackProps: awscdk.StackProps{
			// Env is left nil so CDK resolves the account/region from the
			// default AWS CLI configuration (CDK_DEFAULT_ACCOUNT /
			// CDK_DEFAULT_REGION), which is sufficient for a stack that is
			// only ever deployed manually from a workstation.
		},
	})

	app.Synth(nil)
}
