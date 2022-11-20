#!make

build:
	cd lambda_s3_to_s3 && GOOS=linux GOARCH=amd64 go build -o bin/lambda_s3_to_s3 ./cmd/main.go
