data "aws_region" "current" {}

resource "aws_appsync_api" "dev" {
  name = "appsync-dev"

  event_config {
    // INFO: Valid auth provider types: API_KEY, AWS_IAM, AMAZON_COGNITO_USER_POOLS, OPENID_CONNECT, AWS_LAMBDA.
    auth_provider {
      auth_type = "AWS_IAM"
    }

    auth_provider {
      auth_type = "API_KEY"
    }

    auth_provider {
      auth_type = "AWS_LAMBDA"
      lambda_authorizer_config {
        authorizer_uri                   = aws_lambda_function.authorizer.arn
        authorizer_result_ttl_in_seconds = 300
      }
    }

    connection_auth_mode {
      auth_type = "AWS_IAM"
    }

    connection_auth_mode {
      auth_type = "API_KEY"
    }

    connection_auth_mode {
      auth_type = "AWS_LAMBDA"
    }

    default_publish_auth_mode {
      auth_type = "API_KEY"
    }

    default_subscribe_auth_mode {
      auth_type = "API_KEY"
    }
  }
}


resource "aws_appsync_channel_namespace" "dev" {
  name   = "appsync-dev"
  api_id = aws_appsync_api.dev.api_id
  subscribe_auth_mode {
    auth_type = "AWS_IAM"
  }

  subscribe_auth_mode {
    auth_type = "API_KEY"
  }

  subscribe_auth_mode {
    auth_type = "AWS_LAMBDA"
  }

  publish_auth_mode {
    auth_type = "AWS_IAM"
  }

  publish_auth_mode {
    auth_type = "API_KEY"
  }

  publish_auth_mode {
    auth_type = "AWS_LAMBDA"

  }
}

resource "aws_appsync_api_key" "dev" {
  api_id = aws_appsync_api.dev.api_id
}
