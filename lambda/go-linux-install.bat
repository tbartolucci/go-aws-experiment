set GOARCH=amd64
set GOOS=linux
go build -o main ./thumbnail/main.go
build-lambda-zip.exe -o main.zip main