#!make

build:
	cd lambda_s3_to_s3 && GOOS=linux GOARCH=amd64 go build -o bin/lambda_s3_to_s3 ./cmd/main.go

apply:
	terraform apply -auto-approve -replace="aws_lambda_function.lambda_s3_to_s3"
