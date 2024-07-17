package utils

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

func BuildResponse(body string, statusCode int, headers map[string][]string) events.APIGatewayProxyResponse {
	if statusCode >= 300 {
		return errorResponse(body, statusCode, buildResponseHeaders(headers))
	}
	return dataResponse(body, statusCode, buildResponseHeaders(headers))
}

func dataResponse(body string, statusCode int, headers map[string][]string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:              body,
		StatusCode:        statusCode,
		MultiValueHeaders: headers,
	}
}

func errorResponse(message string, statusCode int, headers map[string][]string) events.APIGatewayProxyResponse {
	type ErrorResponse struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}
	errResp := ErrorResponse{
		Message: message,
		Status:  statusCode,
	}
	body, _ := json.Marshal(errResp)
	return events.APIGatewayProxyResponse{
		Body:              string(body),
		StatusCode:        statusCode,
		MultiValueHeaders: headers,
	}
}

var defaultHeaders = map[string][]string{
	"Cache-Control":             {"no-store"},
	"Content-Security-Policy":   {"frame-ancestors 'none'"},
	"Strict-Transport-Security": {"max-age=31536000; includeSubDomains"},
	"X-Content-Type-Options":    {"nosniff"},
	"X-Frame-Options":           {"DENY"},
	"Referrer-Policy":           {"no-referrer"},
	"Permissions-Policy":        {""},
}

func buildResponseHeaders(customHeaders map[string][]string) map[string][]string {

	for key, value := range customHeaders {
		defaultHeaders[key] = value
	}
	return defaultHeaders
}
