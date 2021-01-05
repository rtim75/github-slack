package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handleRequest(ctx context.Context, event events.SQSEvent) error {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return err
	}
	fmt.Printf("%+v\n", string(eventJSON))
	secret, err := getSecret(os.Getenv("SLACK_TOKEN_SECRET"))
	if err != nil {
		return err
	}

	for _, message := range event.Records {
		err := deleteMessage(message.ReceiptHandle)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func main() {
	lambda.Start(handleRequest)
}
