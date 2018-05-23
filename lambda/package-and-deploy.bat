set GOARCH=amd64
set GOOS=linux
go build -o ./thumbnail/main ./thumbnail/main.go
build-lambda-zip.exe -o main.zip ./thumbnail/main
aws s3api put-object --bucket bitsbybit-pluralsight-photos-lambda --body .\main.zip --key main.zip
del ./thumbnail/main
del main.zip
aws lambda update-function-code --function-name PhotosAppThumbnailFunction --s3-bucket bitsbybit-pluralsight-photos-lambda --s3-key main.zip