package aws

import (
	"context"
	"user/sigmatech/app/constants"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

var awsConfig config.Config

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(constants.Config.AWSConfig.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				constants.Config.AWSConfig.AccessKeyId,
				constants.Config.AWSConfig.SecretAccessKey,
				"",
			),
		),
		// Add any other AWS configuration options here as needed
	)
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}

	awsConfig = cfg
}

// GetAWSConfig returns the AWS SDK config
func GetAWSConfig() config.Config {
	return awsConfig
}
