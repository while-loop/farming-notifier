package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func update(item interface{}, table string) error {
	av, err := dynamodbattribute.MarshalMap(item)
	if err != nil {
		return err
	}

	svc := dynamoSession()

	// Create Item in table and return
	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(table),
	}

	_, err = svc.PutItem(input)
	return err
}

func dynamoSession() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSession())
	svc := dynamodb.New(sess)
	return svc
}

func getUser(username string) (User, error) {
	resp, err := dynamoSession().GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(usersTable),
		Key: map[string]*dynamodb.AttributeValue{
			"username": {
				S: aws.String(username),
			},
		},
	})

	if err != nil {
		return User{}, err
	}

	return UserFromDyn(resp.Item), nil
}