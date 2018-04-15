package go_aws

import (
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/aws"
	"strings"
)

func main() {
	client := s3.New(nil) //use AWS_REGION
	result, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("MyBucketName"),
		Key: aws.String("hello.txt"),
		Body: strings.NewReader("Hello, cloud!"),
	})
}