package main

import (
	"Recipe-API/internal/aws/awsclient"
	"Recipe-API/internal/utils"
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

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
	fileData := request.Body
	if fileData == "" {
		return utils.BuildResponse("No request body supplied!", 400, nil), nil
	}
	fileBytes, err := base64.StdEncoding.DecodeString(fileData)
	if err != nil {
		return utils.BuildResponse("Error decoding base64 string", 500, nil), nil
	}

	s3BucketName, err := GetS3BucketeName(ctx)
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	s3Client, err := awsclient.GetS3Client()
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3BucketName),
		Key:    aws.String(utils.FileName),
		Body:   bytes.NewReader(fileBytes),
	})
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	return utils.BuildResponse("", 201, nil), nil
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
