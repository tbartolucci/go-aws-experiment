package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

const usersTable = "PhotosAppUsers"
const usernameIndex = "Username-index"
const userIdIndex = "UserID-index"

const photosTable = "PhotosAppPhotos"

const followersTable = "PhotosAppFollowers"
const followerIdIndex = "FollowerID-index"

const commentsTable = "PhotosAppComments"

func NewAwsSession() *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("", "go-aws"),
	}))

	return sess
}

func NewDynamoDb() *dynamodb.DynamoDB {
	sess := NewAwsSession()
	svc := dynamodb.New(sess)
	return svc
}
