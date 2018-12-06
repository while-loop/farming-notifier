package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
)

func UpdatePatch(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	patch, err := NewPatch(request.Body)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: "bad patch", StatusCode: http.StatusBadRequest}, nil
	}

	if update(patch, patchesTable) != nil {
		fmt.Printf("failed to update patch: %v\n", err)
		return events.APIGatewayProxyResponse{Body: "server error", StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 204}, nil
}

func main() {
	lambda.Start(UpdatePatch)
}
