resource "aws_lambda_function" "authorizer" {
  function_name    = "${local.name_prefix}-authorizer"
  role             = aws_iam_role.authorizer.arn
  filename         = "../dist/authorizer/function.zip"
  source_code_hash = filebase64sha256("../dist/authorizer/function.zip")
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  architectures    = ["arm64"]

  environment {
    variables = {
      LAMBDA_AUTHORIZER_TOKEN = var.lambda_authorizer_token
    }
  }
}

resource "aws_iam_role" "authorizer" {
  name = "${local.name_prefix}-authorizer"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect    = "Allow"
      Principal = { Service = "lambda.amazonaws.com" }
      Action    = "sts:AssumeRole"
    }]
  })
}

resource "aws_iam_role_policy_attachment" "authorizer-basic-execution-role-attachment" {
  role       = aws_iam_role.authorizer.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_permission" "appsync_invoke_authorizer" {
  statement_id  = "AllowAppSyncInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.authorizer.function_name
  principal     = "appsync.amazonaws.com"
  source_arn    = aws_appsync_api.e2e.api_arn
}
