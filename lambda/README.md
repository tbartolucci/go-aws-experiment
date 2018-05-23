## Build Instructions

[AWS Lambda GO Build](https://github.com/aws/aws-lambda-go)

[AWS GO Lambda SDK](https://docs.aws.amazon.com/lambda/latest/dg/go-programming-model-handler-types.html)

### Build

In Powershell:

`$env:GOOS = "linux"`

`$env:GOARCH = "amd64"`

 `go build -o main main.go`
 
 `~\Go\Bin\build-lambda-zip.exe -o main.zip main`

### Deploy



## Create Lambda Trigger

1. Lambda > Functions > PhotosApp_thumbnail > Triggers > Add trigger
1. Click empty box
1. Select S3
1. Choose bucket (i.e. "pluralsight-photos")
1. Choose event type "Object created (All)"
1. Click Submit

## Edit IAM policy to allow S3 full access

1. IAM > Roles > PhotosApp_lambda_function
1. Permissions > Attach policy
1. Select "AmazonS3FullAccess" checkbox
1. Click "Attach policy"

## Test Lambda Function

1. Upload an image to the bucket in S3

## View CloudWatch Logs

