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

data "aws_iam_policy_document" "iam_policy_lambda_for_s3" {
  statement {
    actions = [
      "s3:GetObject",
      "s3:PutObject",
    ]
    resources = [
      "arn:aws:s3:::${var.bucket_name}/*"
    ]
  }
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = [
      "arn:aws:logs:*:*:*"
    ]
  }
}

resource "aws_iam_policy" "iam_policy_lambda_for_s3" {
  name        = "iam_policy_lambda_for_s3"
  description = "iam_policy_lamda_for_s3 description"

  policy = data.aws_iam_policy_document.iam_policy_lambda_for_s3.json
}

resource "aws_iam_role_policy_attachment" "iam_for_lambda_s3" {
  role       = aws_iam_role.lambda.id
  policy_arn = aws_iam_policy.iam_policy_lambda_for_s3.arn
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

module "sns_topic" {
  source  = "terraform-aws-modules/sns/aws"
  version = "~> 3.0"

  name = "sns-topic-for-lambda"
}

resource "aws_sns_topic_subscription" "user_updates_sqs_target" {
  topic_arn = module.sns_topic.sns_topic_arn
  protocol  = "https"
  endpoint  = "https://global.sns-api.chatbot.amazonaws.com"
}

module "metric_alarms" {
  source  = "terraform-aws-modules/cloudwatch/aws//modules/metric-alarms-by-multiple-dimensions"
  version = "~> 3.0"

  alarm_description = "all-lambda-functions-alarm"
  alarm_name        = "all-lambda-functions-alarm"

  # 以下を参考に「メトリクスを集計する AWSサービス」を指定する
  # https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/aws-services-cloudwatch-metrics.html
  namespace = "AWS/Lambda"

  metric_name = "Errors"

  # メトリクスの統計(集計方法)
  # 以下のいずれかを指定し、集計した値と `threshold` とを `comparison_operator` の比較方法で比較する。
  # Average | Maximum | Minimum | SampleCount | Sum
  statistic = "SampleCount"

  # 閾値
  threshold = 0
  # 閾値と比較する回数
  evaluation_periods = 1

  # `evaluation_periods` 1回あたりの統計が適用される期間（秒単位）
  # 有効な値は、10、30、60、および60の任意の倍数。
  period = 60

  # アラーム状態へ遷移させるか判定する基準
  # `period` * `evaluation_periods` のあいだに `datapoints_to_alarm` 回だけ `threshold` を、
  # `comparison_operator` で指定する比較方法で比較し、TRUE である場合、アラーム状態へ遷移
  # 上記の場合、60秒 * 1 = 60秒のあいだに、閾値である1か、1以上のエラーが検知されたら、アラーム状態へ遷移となる。
  datapoints_to_alarm = 1

  # `threshold` との比較演算子であり、以下のいずれかを指定。
  # GreaterThanThreshold, GreaterThanOrEqualToThreshold, LessThanThreshold, or LessThanOrEqualToThreshold
  comparison_operator = "GreaterThanThreshold"

  # AWS Lambda の場合、以下を参考に入力。
  # https://docs.aws.amazon.com/lambda/latest/dg/monitoring-metrics.html
  # ここで「複数の Lambda」を監視するように指定できる！！
  dimensions = {
    "lambda1" = {
      FunctionName = local.function_name
    }
  }

  alarm_actions = [module.sns_topic.sns_topic_arn]
}
