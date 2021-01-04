package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, event events.SQSEvent) error {
	fmt.Printf("%v+\n", event)
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
