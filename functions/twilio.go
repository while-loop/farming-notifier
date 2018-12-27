package main

import (
	"encoding/json"
	"fmt"
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

func sendText(wg *sync.WaitGroup, user User, message string) {
	defer wg.Done()
	if user.Number == "" || message == "" {
		fmt.Printf("unable to send text. empty number or message %v %v\n", user.Number, message)
		return
	}

	msgData := url.Values{}
	msgData.Set("To", user.Number)
	msgData.Set("From", numberFrom)
	msgData.Set("Body", message)
	msgDataReader := *strings.NewReader(msgData.Encode())

	req, err := http.NewRequest("POST", urlStr, &msgDataReader)

	req.SetBasicAuth(accountSid, authToken)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("unable to twilio request %s: %v\n", user.Username, err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var data map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		if err = decoder.Decode(&data); err != nil {
			fmt.Println("sendText", err)
		} else {
			fmt.Printf("Text sent to %s, %s: %s\n", user.Username, user.Number, data["sid"])
		}
	} else {
		fmt.Printf("twilio unknown response %v: %s\n", msgData, resp.Status)
	}
}
