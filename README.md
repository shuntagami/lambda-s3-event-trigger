# lambda-s3-event-trigger

## What

When zip file uploaded to Amazon S3, extract all the contents from it and upload them. That's it!

## Requirements

[AWS SAM CLI](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html)

or

[Terraform](https://developer.hashicorp.com/terraform/tutorials/aws-get-started/install-cli)

## Build

### SAM CLI

```
$ sam build
```

### Terraform

```
$ cd terraform
$ terraform import aws_s3_bucket.shuntagami-demo-data shuntagami-demo-data
$ cd ..
$ make build
$ terraform plan
```

## Deploy

### SAM CLI

```
$ sam deploy
```

### Terraform

```
$ cd terraform
$ terraform plan
$ terraform apply
```

If you modify function, run below

```
$ make build && cd terraform
$ terraform apply -replace="aws_lambda_function.lambda_s3_to_s3"
```

## Cleanup

```
$ aws s3 rm s3://shuntagami-demo-data/ --recursive
```

### SAM CLI

```
$ aws cloudformation delete-stack --stack-name lambda-s3-to-s3-demo
```

### Terraform

```
$ terraform destroy
```

## License

MIT

