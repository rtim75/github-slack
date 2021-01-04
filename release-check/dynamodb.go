package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
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

func updateRepositoryRelease(repository string, tag string, updated int64) error {
	svc := dynamodb.New(session.New())

	cond := expression.Or(expression.AttributeNotExists(expression.Name("lastUpdate")), expression.Name("lastUpdate").LessThan(expression.Value(updated)))
	// cond := expression.AttributeNotExists(expression.Name("lastUpdate"))
	expr, err := expression.NewBuilder().WithCondition(cond).Build()
	if err != nil {
		fmt.Printf("Failed to build a condition: %v\n", err)
		return err
	}
	fmt.Println(expr.Condition(), expr.Names(), expr.Values())

	input := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"repository": {
				S: aws.String(repository),
			},
			"lastTag": {
				S: aws.String(tag),
			},
			"lastUpdate": {
				N: aws.String(strconv.FormatInt(updated, 10)),
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
				return nil
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
		return err
	}

	return nil

}
