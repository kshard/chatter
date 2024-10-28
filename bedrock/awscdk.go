package bedrock

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awsbedrock"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

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

func NewTitanTextLiteV1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("TitanTextLiteV1"),
		awsbedrock.FoundationModelIdentifier_AMAZON_TITAN_TEXT_LITE_V1(),
	)
}

func NewTitanTextExpressV1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("TitanTextExpressV1"),
		awsbedrock.FoundationModelIdentifier_AMAZON_TITAN_TEXT_EXPRESS_V1_0_8K(),
	)
}

func NewTitanTextPremierV1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("TitanTextPremierV1"),
		awsbedrock.FoundationModelIdentifier_AMAZON_TITAN_TEXT_PREMIER_V1(),
	)
}

func NewMetaLlama30B8V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama30B8V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_8B_INSTRUCT_V1(),
	)
}

func NewMetaLlama30B70V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama30B70V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_70_INSTRUCT_V1(),
	)
}

func NewMetaLlama31B8V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama31B8V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_1_8B_INSTRUCT_V1(),
	)
}

func NewMetaLlama31B70V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama31B70V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_1_70_INSTRUCT_V1(),
	)
}

func NewMetaLlama31B405V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama31B405V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_1_405_INSTRUCT_V1(),
	)
}

func NewMetaLlama32B1V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama31B1V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_2_1B_INSTRUCT_V1(),
	)
}

func NewMetaLlama32B3V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama31B3V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_2_3B_INSTRUCT_V1(),
	)
}

func NewMetaLlama32B11V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama31B11V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_2_11B_INSTRUCT_V1(),
	)
}

func NewMetaLlama32B90V1(scope constructs.Construct) *FoundationModel {
	return NewFoundationModel(scope, jsii.String("MetaLlama31B90V1"),
		awsbedrock.FoundationModelIdentifier_META_LLAMA_3_2_90B_INSTRUCT_V1(),
	)
}
