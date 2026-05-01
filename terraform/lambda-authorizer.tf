resource "aws_lambda_function" "authorizer" {
  function_name    = "appsync-lambda-authorizer"
  role             = aws_iam_role.authorizer.arn
  filename         = "../dist/authorizer/function.zip"
  source_code_hash = filebase64sha256("../dist/authorizer/function.zip")
  handler          = "bootstrap"
  runtime          = "provided.al2"
  architectures    = ["arm64"]
}

resource "aws_iam_role" "authorizer" {
  name = "authorizer"

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

resource "aws_iam_role_policy" "authorizer-role-attachment" {
  role = aws_iam_role.authorizer.name
  name = "authorizer-role-permissions-attachment"
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
      ],
      Resource = ""
    }]
  })
}
