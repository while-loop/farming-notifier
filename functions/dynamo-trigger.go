package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"sync"
	"time"
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
		user, err  := getUser(username)
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
		go sendText(&wg, user, message)
	}

	wg.Wait()
	return nil
}

func main() {
	lambda.Start(DynamoTrigger)
}
