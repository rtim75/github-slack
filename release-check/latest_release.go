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
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/mmcdole/gofeed"
)

type repositoryLatestRelease struct {
	Repository   string    `json:"repository"`
	LatestTag    string    `json:"latestTag"`
	LatestUpdate time.Time `json:"latestUpdate"`
}

func getRepositories() ([]repositoryLatestRelease, error) {
	svc := dynamodb.New(session.New())
	input := &dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("REPOSITORIES_TABLE")),
	}
	repositories := []repositoryLatestRelease{}

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

	for _, repository := range result.Items {
		latestTag := "0.0.0"
		if _, ok := repository["lastTag"]; ok {
			latestTag = *repository["lastTag"].S
		}

		lastUpdate := time.Unix(0, 0)
		if _, ok := repository["lastUpdate"]; ok {
			tsParsed, err := strconv.ParseInt(*repository["lastUpdate"].N, 10, 64)
			if err != nil {
				fmt.Printf("Failed to parse timestamp %v: %v\n", *repository["lastUpdate"].N, err)
			}
			lastUpdate = time.Unix(tsParsed, 0)
		}

		repositories = append(repositories, repositoryLatestRelease{
			Repository:   *repository["repository"].S,
			LatestTag:    latestTag,
			LatestUpdate: lastUpdate,
		})
	}

	return repositories, nil
}

func (release *repositoryLatestRelease) save() (bool, error) {
	svc := dynamodb.New(session.New())

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
		TableName: aws.String(os.Getenv("REPOSITORIES_TABLE")),
	}

	_, err := svc.PutItem(input)
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

func (release *repositoryLatestRelease) getLatest() (bool, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL("https://" + release.Repository + "/releases.atom")
	if err != nil {
		fmt.Printf("Failed to parse feed for %v: %v", release.Repository, err)
		return false, err
	}

	if release.LatestTag != feed.Items[0].Title {
		release.LatestTag = feed.Items[0].Title
		release.LatestUpdate = *feed.Items[0].UpdatedParsed

		return true, nil
	}

	return false, nil

}
