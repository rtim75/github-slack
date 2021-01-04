package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/mmcdole/gofeed"
)

func getRepositories() ([]string, error) {
	svc := dynamodb.New(session.New())
	input := &dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("REPOSITORIES_TABLE")),
	}
	repositories := []string{}

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
		repositories = append(repositories, *repository["repository"].S)
	}

	return repositories, nil
}

func getLatestRelease(repository string) (*repositoryLatestRelease, error) {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL("https://" + repository + "/releases.atom")
	if err != nil {
		fmt.Printf("Failed to parse feed for %v: %v", repository, err)
		return nil, err
	}

	return &repositoryLatestRelease{
		Repository:   repository,
		LatestTag:    feed.Items[0].Title,
		LatestUpdate: *feed.Items[0].UpdatedParsed,
	}, nil
}
