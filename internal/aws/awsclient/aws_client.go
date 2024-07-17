package awsclient

import (
	"Recipe-API/internal/aws/awsconfig"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func GetSsmClient() (*ssm.Client, error) {
	client, err := createClient(func(cfg aws.Config) interface{} {
		return ssm.NewFromConfig(cfg)
	})
	if err != nil {
		return nil, err
	}
	return client.(*ssm.Client), nil
}

func GetS3Client() (*s3.Client, error) {
	client, err := createClient(func(cfg aws.Config) interface{} {
		return s3.NewFromConfig(cfg)
	})
	if err != nil {
		return nil, err
	}
	return client.(*s3.Client), nil
}

func createClient(factory func(cfg aws.Config) interface{}) (interface{}, error) {
	awsConf, err := awsconfig.GetAwsConfig()
	if err != nil {
		return nil, err
	}
	client := factory(awsConf)
	return client, nil
}
