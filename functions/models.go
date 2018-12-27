package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"strings"
)

type User struct {
	Username string `json:"username"`
	Number   string `json:"number"`
}

type Patch struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Region   string `json:"region"`
	State    string `json:"state"`
	Produce  string `json:"produce"`
	Patch    string `json:"patch"`
	Type     string `json:"type"`
	TTL      int64  `json:"ttl"`
}

func NewPatch(item string) (Patch, error) {
	p := Patch{}
	err := json.Unmarshal([]byte(item), &p)
	if err != nil {
		fmt.Printf("failed to unmarshal patch: %v\n", err)
		return Patch{}, err
	}

	p.Region = strings.Title(p.Region)
	p.State = strings.Title(strings.ToLower(strings.Replace(p.State, "_", " ", -1)))
	p.Produce = strings.Title(strings.ToLower(strings.Replace(p.Produce, "_", " ", -1)))
	p.Patch = strings.Title(strings.ToLower(p.Patch))
	p.Type = strings.Title(strings.ToLower(strings.Replace(p.Type, "_", " ", -1)))
	p.ID = fmt.Sprintf("%s.%s.%s.%s", p.Username, p.Region, p.Patch, p.Type)
	return p, nil
}

func FromDyn(av map[string]events.DynamoDBAttributeValue) Patch {
	ttl, err := av["ttl"].Integer()
	if err != nil {
		fmt.Printf("err getting ttl: %v", err)
		ttl = 0
	}

	return Patch{
		ID:       safeString(av["id"]),
		Username: safeString(av["username"]),
		Region:   safeString(av["region"]),
		Patch:    safeString(av["patch"]),
		Type:     safeString(av["type"]),
		State:    safeString(av["state"]),
		Produce:  safeString(av["produce"]),
		TTL:      ttl,
	}
}

func UserFromDyn(av map[string]*dynamodb.AttributeValue) User {
	return User{
		Username: safeAV(av["username"]),
		Number:   safeAV(av["number"]),
	}
}

func safeAV(av *dynamodb.AttributeValue) string {
	if av == nil  || av.S == nil {
		return ""
	}

	return *av.S
}

func safeString(av events.DynamoDBAttributeValue) string {
	if av.IsNull() {
		return ""
	}

	return av.String()
}

const (
	removeEvent  = "REMOVE"
	usersTable   = "Users"
	patchesTable = "Patches"
)
