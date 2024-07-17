package awsconfig

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
)

func GetAwsConfig() (aws.Config, error) {
	ctx := context.Background()
	awsConfig, err := config.LoadDefaultConfig(ctx,
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 5)
		}),
	)
	if err != nil {
		return aws.Config{}, err
	}
	return awsConfig, nil
}
