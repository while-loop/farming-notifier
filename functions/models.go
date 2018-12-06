package main

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

type User struct {
	Username string `json:"username"`
	Number   string `json:"number"`
}

type Patch struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Region   string `json:"region"`
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
	p.ID = fmt.Sprintf("%s.%s.%s.%s", p.Username, p.Region, p.Patch, p.Type)
	return p, nil
}

func FromDyn(av map[string]events.DynamoDBAttributeValue) Patch {
	ttl, err := av["ttl"].Integer()
	if err != nil {
		ttl = 0
	}

	return Patch{
		ID:       av["id"].String(),
		Username: av["username"].String(),
		Region:   av["region"].String(),
		Patch:    av["patch"].String(),
		Type:     av["type"].String(),
		TTL:      ttl,
	}
}

const (
	removeEvent  = "REMOVE"
	usersTable   = "Users"
	patchesTable = "Patches"
)
