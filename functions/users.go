package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
)

func UpdateUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	user := &User{}
	err := json.Unmarshal([]byte(request.Body), user)
	if err != nil {
		fmt.Printf("failed to unmarshal user: %v\n", err)
		return events.APIGatewayProxyResponse{Body: "bad user", StatusCode: http.StatusBadRequest}, nil
	}

	if update(user, usersTable) != nil {
		fmt.Printf("failed to update user: %v\n", err)
		return events.APIGatewayProxyResponse{Body: "server error", StatusCode: http.StatusInternalServerError}, nil
	}

	return events.APIGatewayProxyResponse{StatusCode: 204}, nil
}

func main() {
	lambda.Start(UpdateUser)
}
