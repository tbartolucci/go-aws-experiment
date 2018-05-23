package main

import (
	"bytes"
	"image/jpeg"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/nfnt/resize"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"context"
	"github.com/aws/aws-lambda-go/lambda"
)

type ImageEvent struct {
	bucket string
	key string
}

func HandleRequest(ctx context.Context, image ImageEvent) (string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewSharedCredentials("", "go-aws"),
	}))

	bucket := image.bucket
	key := image.key

	log.Printf("Bucket: %s", bucket)
	log.Printf("Key: %s", key)

	// Prevent recursive Lambda trigger

	if strings.Contains(key, "/thumb/") {
		return "", nil
	}

	log.Printf("Fetching s3://%v/%v", bucket, key)

	buff := &aws.WriteAtBuffer{}
	s3dl := s3manager.NewDownloader(sess)
	_, err := s3dl.Download(buff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		log.Printf("Could not download from S3: %v", err)
		return "Could not download from S3", err
	}

	log.Printf("Decoding image: %v bytes", len(buff.Bytes()))

	imageBytes := buff.Bytes()
	reader := bytes.NewReader(imageBytes)

	img, err := jpeg.Decode(reader)
	if err != nil {
		log.Printf("bad response: %s", err)
		return "bad response", err
	}

	log.Printf("Generating thumbnail")
	thumbnail := resize.Thumbnail(600, 600, img, resize.Lanczos3)

	if thumbnail == nil {
		log.Printf("resize.Thumbnail returned nil")
		return "resize.Thumbnail return nil", nil
	}

	log.Printf("Encoding image for upload to S3")
	buf := new(bytes.Buffer)
	err = jpeg.Encode(buf, thumbnail, nil)

	if err != nil {
		log.Printf("JPEG encoding error: %v", err)
	}

	// Filename: e5f97749-5d2f-4770-89ce-5d68b1a90f7b/filename.jpg
	// Thumbnail: e5f97749-5d2f-4770-89ce-5d68b1a90f7b/thumb/filename.jpg

	thumbkey := strings.Replace(key, "/", "/thumb/", -1)

	log.Printf("Preparing S3 object: %s", thumbkey)

	uploader := s3manager.NewUploader(sess)
	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   bytes.NewReader(buf.Bytes()),
		Bucket: aws.String(bucket),
		Key:    aws.String(thumbkey),
	})

	if err != nil {
		log.Printf("Failed to upload: %v", err)
	}

	log.Printf("Successfully uploaded to: %v", result.Location)

	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
