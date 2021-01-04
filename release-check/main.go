package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, event events.CloudWatchEvent) error {
	repos, err := getRepositories()
	if err != nil {
		os.Exit(1)
	}

	for i := range repos {
		changed, err := repos[i].getLatest()
		if err != nil {
			os.Exit(1)
		}

		if changed {
			changed, err := repos[i].save()
			if err != nil {
				os.Exit(1)
			}
			if changed {
				err = repos[i].notify()
				if err != nil {
					os.Exit(1)
				}
			}
		}
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
