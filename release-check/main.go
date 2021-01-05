package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, event events.CloudWatchEvent) error {
	repos, err := getRepositories()
	if err != nil {
		return err
	}

	for i := range repos {
		changed, err := repos[i].getLatest()
		if err != nil {
			return err
		}

		if changed {
			changed, err := repos[i].save()
			if err != nil {
				return err
			}
			if changed {
				err = repos[i].notify()
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
