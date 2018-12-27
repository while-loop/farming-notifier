package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"net/http"
	"sync"
)

func UpdateUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Println("Received body: ", request.Body)

	user := &User{}
	err := json.Unmarshal([]byte(request.Body), user)
	if err != nil {
		fmt.Printf("failed to unmarshal user: %v\n", err)
		return events.APIGatewayProxyResponse{Body: "bad user", StatusCode: http.StatusBadRequest}, nil
	}

	exists := true

	oldUser, err := getUser(user.Username)
	if  err != nil || oldUser.Username != user.Username {
		log.Println("err getting old user", err)
		exists = false
	}

	if err = update(user, usersTable); err != nil {
		fmt.Printf("failed to update user: %v\n", err)
		return events.APIGatewayProxyResponse{Body: "server error", StatusCode: http.StatusInternalServerError}, nil
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)
	if user.Number != "" && (oldUser.Number == "" || !exists) {
		sendText(wg, *user, fmt.Sprintf("%s has subscribed to OSRS Notifier.\n To stop future texts, remove your number from RuneLite.\n Happy 'Scaping!", user.Username))
	} else {
		sendText(wg, oldUser, fmt.Sprintf("%s has unsubscribed from OSRS Notifier", user.Username))
	}

	return events.APIGatewayProxyResponse{StatusCode: 204}, nil
}

func main() {
	lambda.Start(UpdateUser)
}
