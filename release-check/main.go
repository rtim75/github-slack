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
		release, err := getLatestRelease(repos[i])
		if err != nil {
			os.Exit(1)
		}

		err = updateRepositoryRelease(release.name, release.latestTag, release.released.Unix())
		if err != nil {
			os.Exit(1)
		}
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
