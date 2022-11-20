locals {
  function_name = "lambda_s3_to_s3"
}
resource "aws_s3_bucket" "shuntagami-demo-data" {
  bucket = var.bucket_name
}

data "archive_file" "lambda_functions" {
  type             = "zip"
  source_file      = "../lambda_s3_to_s3/bin/lambda_s3_to_s3"
  output_file_mode = "0666"
  output_path      = "${path.root}/tmp/lambda_s3_to_s3.zip"
}

# codes for lambda functions
resource "aws_iam_role" "lambda" {
  name = "iam_for_lambda"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Action" : "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy" "lambda" {
  name = "s3_policy"
  role = aws_iam_role.lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
        ]
        Resource = [
          "arn:aws:s3:::${var.bucket_name}/*",
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
        ]
        Resource = [
          "arn:aws:logs:*:*:*",
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogStream",
          "logs:PutLogEvents",
        ]
        Resource = [
          "arn:aws:logs:*:*:*",
        ]
      },
    ]
  })
}

resource "aws_lambda_function" "lambda_s3_to_s3" {
  filename      = "${path.root}/tmp/lambda_s3_to_s3.zip"
  function_name = local.function_name
  role          = aws_iam_role.lambda.arn
  handler       = local.function_name # binary file's name here
  runtime       = "go1.x"
  timeout       = 60
}

resource "aws_s3_bucket_notification" "incoming" {
  bucket = aws_s3_bucket.shuntagami-demo-data.id

  lambda_function {
    lambda_function_arn = aws_lambda_function.lambda_s3_to_s3.arn
    events              = ["s3:ObjectCreated:*"]
    filter_suffix       = ".zip"
  }

  depends_on = [aws_lambda_permission.s3_permission_to_trigger_lambda]
}

resource "aws_lambda_permission" "s3_permission_to_trigger_lambda" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda_s3_to_s3.arn
  principal     = "s3.amazonaws.com"
  source_arn    = aws_s3_bucket.shuntagami-demo-data.arn
}
