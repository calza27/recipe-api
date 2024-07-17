package main

import (
	"Recipe-API/internal/aws/awsclient"
	"Recipe-API/internal/utils"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fileVersion := request.PathParameters["version"]
	if fileVersion == "" {
		return utils.BuildResponse("No file version supplied in path!", 400, nil), nil
	}
	s3BucketName, err := GetS3BucketeName(ctx)
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	fileLifespan, err := GetFileUrlLifeSpan(ctx)
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	s3Client, err := awsclient.GetS3Client()
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	_, err = s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket:    aws.String(s3BucketName),
		Key:       aws.String(utils.FileName),
		VersionId: aws.String(fileVersion),
	})
	if err != nil {
		return utils.BuildResponse(fmt.Sprintf("version %s not found for file %s", fileVersion, utils.FileName), 404, nil), nil
	}
	presignClient := s3.NewPresignClient(s3Client)
	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:    aws.String(s3BucketName),
		Key:       aws.String(utils.FileName),
		VersionId: aws.String(fileVersion),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Duration(fileLifespan)
	})
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	return utils.BuildResponse(req.URL, 200, nil), nil
}

func GetS3BucketeName(ctx context.Context) (string, error) {
	ssmClient, err := awsclient.GetSsmClient()
	if err != nil {
		return "", fmt.Errorf("error initializing connection to SSM: %w", err)
	}
	params := &ssm.GetParameterInput{
		Name: aws.String("/recipe-api-processor/s3/name"),
	}
	s3Name, err := ssmClient.GetParameter(ctx, params)
	if err != nil {
		return "", fmt.Errorf("error when getting parameter: %w", err)
	}
	return *s3Name.Parameter.Value, nil
}

func GetFileUrlLifeSpan(ctx context.Context) (time.Duration, error) {
	ssmClient, err := awsclient.GetSsmClient()
	if err != nil {
		return 0, fmt.Errorf("error initializing connection to SSM: %w", err)
	}
	params := &ssm.GetParameterInput{
		Name: aws.String("/recipe-api-processor/s3/url-duration"),
	}
	duration, err := ssmClient.GetParameter(ctx, params)
	if err != nil {
		return 0, fmt.Errorf("error when getting parameter: %w", err)
	}
	return time.ParseDuration(*duration.Parameter.Value)
}
