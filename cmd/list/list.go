package main

import (
	"Recipe-API/internal/aws/awsclient"
	"Recipe-API/internal/models"
	"Recipe-API/internal/utils"
	"context"
	"encoding/json"
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
	bucektName, err := GetS3BucketeName(ctx)
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	s3Client, err := awsclient.GetS3Client()
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	resp, err := s3Client.ListObjectVersions(ctx, &s3.ListObjectVersionsInput{
		Bucket: aws.String(bucektName),
		Prefix: aws.String(utils.FileName),
	})
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}

	files := []models.FileData{}
	for _, obk := range resp.Versions {
		fileData := models.FileData{
			Version: *obk.VersionId,
			Date:    obk.LastModified.Format(time.DateTime),
			PdfFile: *obk.Key,
		}
		files = append(files, fileData)
	}
	data, err := json.Marshal(files)
	if err != nil {
		return utils.BuildResponse(err.Error(), 500, nil), nil
	}
	return utils.BuildResponse(string(data), 200, nil), nil
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
