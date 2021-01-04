package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type repositoryLatestRelease struct {
	Repository   string    `json:"repository"`
	LatestTag    string    `json:"latestTag"`
	LatestUpdate time.Time `json:"latestUpdate"`
}

func (release *repositoryLatestRelease) save() (bool, error) {
	svc := dynamodb.New(session.New())

	cond := expression.Or(expression.AttributeNotExists(expression.Name("lastUpdate")), expression.Name("lastUpdate").LessThan(expression.Value(release.LatestUpdate)))
	// cond := expression.AttributeNotExists(expression.Name("lastUpdate"))
	expr, err := expression.NewBuilder().WithCondition(cond).Build()
	if err != nil {
		fmt.Printf("Failed to build a condition: %v\n", err)
		return false, err
	}

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"repository": {
				S: aws.String(release.Repository),
			},
			"lastTag": {
				S: aws.String(release.LatestTag),
			},
			"lastUpdate": {
				N: aws.String(strconv.FormatInt(release.LatestUpdate.Unix(), 10)),
			},
		},
		ConditionExpression:       expr.Condition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		TableName:                 aws.String(os.Getenv("REPOSITORIES_TABLE")),
	}

	_, err = svc.PutItem(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return false, nil
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Printf("Failed to put item: %v %v\n", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Printf("Failed to put item: %v %v\n", dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeItemCollectionSizeLimitExceededException:
				fmt.Printf("Failed to put item: %v %v\n", dynamodb.ErrCodeItemCollectionSizeLimitExceededException, aerr.Error())
			case dynamodb.ErrCodeTransactionConflictException:
				fmt.Printf("Failed to put item: %v %v\n", dynamodb.ErrCodeTransactionConflictException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Printf("Failed to put item: %v %v\n", dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Printf("Failed to put item: %v %v\n", dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Printf("Failed to put item: %v\n", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Printf("Something not predicted happened: %v\n", err.Error())
		}
		return false, err
	}
	return true, nil
}

func (release *repositoryLatestRelease) notify() error {
	message, err := json.Marshal(release)
	if err != nil {
		fmt.Printf("Failed to marshal json for release: %v\n", err)
	}

	svc := sqs.New(session.New())
	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(message)),
		QueueUrl:    aws.String(os.Getenv("RELEASES_QUEUE_URL")),
	}

	_, err = svc.SendMessage(input)
	if err != nil {
		fmt.Printf("Failed to send message: %v\n", err)
		return err
	}

	return nil
}
