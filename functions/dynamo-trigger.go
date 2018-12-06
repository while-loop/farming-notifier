package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	client = &http.Client{
		Timeout: 10 * time.Second,
	}
	accountSid = os.Getenv("TWILIO_ACCOUNT_SID")
	numberFrom = os.Getenv("TWILIO_NUMBER")
	authToken  = os.Getenv("TWILIO_AUTH_TOKEN")
	urlStr     = "https://api.twilio.com/2010-04-01/Accounts/" + accountSid + "/Messages.json"
)

func DynamoTrigger(e events.DynamoDBEvent) error {
	fmt.Println("got events!")
	bs, err := json.Marshal(e)
	if err != nil {
		fmt.Printf("got err marshalling dynamo event: %v\n", err)
		return err
	}

	fmt.Println(string(bs))

	expires := map[string][]Patch{}

	for _, event := range e.Records {
		if event.EventName != removeEvent {
			continue
		}

		patch := event.Change.OldImage
		if patch == nil {
			fmt.Println("no new image found for event", event)
			continue
		}

		username := patch["username"].String()
		_, exists := expires[username]
		if !exists {
			expires[username] = []Patch{}
		}

		p := FromDyn(patch)
		fmt.Printf("got expired patch: %v\n", p)
		expires[username] = append(expires[username], p)
	}

	var wg sync.WaitGroup
	for username, patches := range expires {
		user, err := dynamoSession().GetItem(&dynamodb.GetItemInput{
			TableName: aws.String(usersTable),
			Key: map[string]*dynamodb.AttributeValue{
				"username": {
					S: aws.String(username),
				},
			},
		})

		if err != nil {
			fmt.Printf("got error looking up user number %v: %v\n", username, err)
			continue
		}

		var regions []string
		for _, p := range patches {
			regions = append(regions, p.Region+"/"+p.Type)
		}

		wg.Add(1)
		sendText(&wg, username, user.Item["number"].S, fmt.Sprintf("Your patches on %s are ready to harvest! %v", username, regions))
	}

	wg.Wait()
	return nil
}

func sendText(wg *sync.WaitGroup, username string, number *string, message string) {
	defer wg.Done()
	if number == nil || message == "" {
		fmt.Printf("unable to send text. empty number or message %v %v\n", number, message)
		return
	}

	msgData := url.Values{}
	msgData.Set("To", *number)
	msgData.Set("From", numberFrom)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())

	req, err := http.NewRequest("POST", urlStr, &msgDataReader)

	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unable to twilio request %s: %v\n", username, err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&data); err != nil {
			fmt.Println("sendText", err)
		} else {
			fmt.Printf("Text sent to %s, %s: %s\n", username, *number, data["sid"])
		}
	} else {
		fmt.Printf("unknown response %s: %s\n", username, resp.Status)
	}
}

func main() {
	lambda.Start(DynamoTrigger)
}
