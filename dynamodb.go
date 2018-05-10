package main

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/aws"
)

type DynamoDB struct {
	svc *dynamodb.DynamoDB
}

func NewDynamoDb() DynamoDB {
	sess := NewAwsSession()
	svc := dynamodb.New(sess)
	return DynamoDB{svc}
}

func (d DynamoDB) QueryWhereFieldEquals(tableName string, field string, value string) (*dynamodb.QueryOutput, error) {
	queryInput := &dynamodb.QueryInput{
		TableName: aws.String(tableName),
		Limit: aws.Int64(1),
		KeyConditions: map[string]*dynamodb.Condition{
			field: {
				ComparisonOperator: aws.String("EQ"),
				AttributeValueList: []*dynamodb.AttributeValue{
					{
						S: aws.String(value),
					},
				},
			},
		},
	}

	return d.svc.Query(queryInput)
}

func (d DynamoDB) AddItem()
{

}