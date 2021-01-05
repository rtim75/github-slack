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
	"github.com/mmcdole/gofeed"
)

type repository struct {
	Name  string    `json:"name"`
	Tag   string    `json:"tag"`
	TagTS time.Time `json:"tag_ts"`
}

func (release *repository) getLatest() (bool, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL("https://" + release.Name + "/releases.atom")
	if err != nil {
		fmt.Printf("Failed to parse feed for %v: %v", release.Name, err)
		return false, err
	}

	if release.Tag != feed.Items[0].Title {
		release.Tag = feed.Items[0].Title
		release.TagTS = *feed.Items[0].UpdatedParsed

		return true, nil
	}

	return false, nil
}

func (release *repository) save() (bool, error) {
	svc := dynamodb.New(session.New())

	update := expression.Set(expression.Name("Tag"), expression.Value(release.Tag)).Set(expression.Name("TagTS"), expression.Value(release.TagTS.Unix()))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	updateInput := &dynamodb.UpdateItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"Repository": {
				S: aws.String(release.Name),
			},
		},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		TableName:                 aws.String(os.Getenv("REPOSITORIES_TABLE")),
	}

	_, err = svc.UpdateItem(updateInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				fmt.Printf("Failed to put item: %v %v\n", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
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

func (release *repository) notify() error {
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

func getRepositories() ([]repository, error) {
	svc := dynamodb.New(session.New())
	input := &dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("REPOSITORIES_TABLE")),
	}
	repositories := []repository{}

	result, err := svc.Scan(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case dynamodb.ErrCodeProvisionedThroughputExceededException:
				fmt.Printf("Failed to scan the repositories table: %v %v\n", dynamodb.ErrCodeProvisionedThroughputExceededException, aerr.Error())
			case dynamodb.ErrCodeResourceNotFoundException:
				fmt.Printf("Failed to scan the repositories table: %v %v\n", dynamodb.ErrCodeResourceNotFoundException, aerr.Error())
			case dynamodb.ErrCodeRequestLimitExceeded:
				fmt.Printf("Failed to scan the repositories table: %v %v\n", dynamodb.ErrCodeRequestLimitExceeded, aerr.Error())
			case dynamodb.ErrCodeInternalServerError:
				fmt.Printf("Failed to scan the repositories table: %v %v\n", dynamodb.ErrCodeInternalServerError, aerr.Error())
			default:
				fmt.Printf("Failed to scan the repositories table: %v\n", aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Printf("Something not predicted happened: %v\n", err.Error())
		}
		return nil, err
	}

	for _, repo := range result.Items {
		latestTag := "0.0.0"
		if _, ok := repo["Tag"]; ok {
			latestTag = *repo["Tag"].S
		}

		lastUpdate := time.Unix(0, 0)
		fmt.Printf("%+v", repo)
		if _, ok := repo["TagTS"]; ok {
			tsParsed, err := strconv.ParseInt(*repo["TagTS"].N, 10, 64)
			if err != nil {
				fmt.Printf("Failed to parse timestamp %v: %v\n", *repo["TagTS"].N, err)
			}
			lastUpdate = time.Unix(tsParsed, 0)
		}

		repositories = append(repositories, repository{
			Name:  *repo["Repository"].S,
			Tag:   latestTag,
			TagTS: lastUpdate,
		})
	}

	return repositories, nil
}
