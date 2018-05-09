package main

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
)

type AwsSession struct {
	sess *session.Session
}

func NewAwsSession() *session.Session {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("", "default"),
	}))

	return sess
}