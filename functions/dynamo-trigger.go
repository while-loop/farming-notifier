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

	// map[username][type] = []Patch
	expires := map[string]map[string][]Patch{}

	for _, event := range e.Records {
		if event.EventName != removeEvent {
			continue
		}

		p := event.Change.OldImage
		if p == nil {
			fmt.Println("no new image found for event", event)
			continue
		}

		patch := FromDyn(p)
		_, exists := expires[patch.Username]
		if !exists {
			expires[patch.Username] = map[string][]Patch{}
		}

		if patch.TTL == 0 || patch.TTL >= time.Now().Unix() {
			fmt.Printf("got ttl that was in the future %v\n", patch)
			continue
		}

		fmt.Printf("got expired patch: %v\n", patch)
		expires[patch.Username][patch.Type] = append(expires[patch.Username][patch.Type], patch)
	}

	var wg sync.WaitGroup
	for username, types := range expires {
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

		message := fmt.Sprintf("Your patches on %s are ready to harvest!\n", username)
		for patchType, patches := range types {
			var regions []string
			for _, p := range patches {
				regions = append(regions, p.Region)
			}

			message += fmt.Sprintf("\n%s: %v", patchType, regions)
		}

		wg.Add(1)
		go sendText(&wg, username, user.Item["number"].S, message)
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
