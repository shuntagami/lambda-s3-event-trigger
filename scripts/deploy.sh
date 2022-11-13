#!/usr/bin/env bash

# Build, package, deploy sam template containing all resources

#set -x

export S3_BUCKET_BUILD="shuntagami-sam-demo"
export SAM_CLI_TELEMETRY=0

# validate
# sam validate -t template.yaml
# aws cloudformation validate-template \
#   --template-url https://shuntagami-sam-demo.s3.ap-northeast-1.amazonaws.com/template.yaml

# build, package, deploy
time sam build

time sam package \
  --output-template-file packaged.yaml \
  --s3-bucket "${S3_BUCKET_BUILD}"

time sam deploy \
  -t packaged.yaml \
  --stack-name lambda-s3-to-s3-demo \
  --capabilities CAPABILITY_IAM \
  --debug

# catch error - cloudformation returns code 255
# (exit non-zero) if stack exists and no updates
if [ "$?" -eq 255 ]
then
    echo "No changes to deploy."
    true
fi
