resource "aws_appsync_api" "e2e" {
  name = local.name_prefix


  event_config {
    auth_provider {
      auth_type = "API_KEY"
    }

    auth_provider {
      auth_type = "AWS_IAM"
    }

    auth_provider {
      auth_type = "AWS_LAMBDA"

      lambda_authorizer_config {
        authorizer_uri                   = aws_lambda_function.authorizer.arn
        authorizer_result_ttl_in_seconds = 0
      }
    }

    auth_provider {
      auth_type = "AMAZON_COGNITO_USER_POOLS"

      cognito_config {
        user_pool_id = aws_cognito_user_pool.e2e.id
        aws_region   = var.aws_region
      }
    }

    auth_provider {
      auth_type = "OPENID_CONNECT"

      openid_connect_config {
        issuer   = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.e2e.id}"
        auth_ttl = 3600000
        iat_ttl  = 3600000
      }
    }

    connection_auth_mode {
      auth_type = "API_KEY"
    }

    connection_auth_mode {
      auth_type = "AWS_IAM"
    }

    connection_auth_mode {
      auth_type = "AWS_LAMBDA"
    }

    connection_auth_mode {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
    }

    connection_auth_mode {
      auth_type = "OPENID_CONNECT"
    }

    default_publish_auth_mode {
      auth_type = "API_KEY"
    }

    default_subscribe_auth_mode {
      auth_type = "API_KEY"
    }
  }
}
resource "aws_appsync_api_key" "e2e" {
  api_id = aws_appsync_api.e2e.api_id
}

resource "aws_appsync_channel_namespace" "api_key" {
  api_id = aws_appsync_api.e2e.api_id
  name   = "api-key-e2e"

  publish_auth_mode {
    auth_type = "API_KEY"
  }

  subscribe_auth_mode {
    auth_type = "API_KEY"
  }
}

resource "aws_appsync_channel_namespace" "iam" {
  api_id = aws_appsync_api.e2e.api_id
  name   = "iam-e2e"

  publish_auth_mode {
    auth_type = "AWS_IAM"
  }

  subscribe_auth_mode {
    auth_type = "AWS_IAM"
  }
}

resource "aws_appsync_channel_namespace" "lambda" {
  api_id = aws_appsync_api.e2e.api_id
  name   = "lambda-e2e"

  publish_auth_mode {
    auth_type = "AWS_LAMBDA"
  }

  subscribe_auth_mode {
    auth_type = "AWS_LAMBDA"
  }
}

resource "aws_appsync_channel_namespace" "cognito" {
  api_id = aws_appsync_api.e2e.api_id
  name   = "cognito-e2e"

  publish_auth_mode {
    auth_type = "AMAZON_COGNITO_USER_POOLS"
  }

  subscribe_auth_mode {
    auth_type = "AMAZON_COGNITO_USER_POOLS"
  }
}

resource "aws_appsync_channel_namespace" "oidc" {
  api_id = aws_appsync_api.e2e.api_id
  name   = "oidc-e2e"

  publish_auth_mode {
    auth_type = "OPENID_CONNECT"
  }

  subscribe_auth_mode {
    auth_type = "OPENID_CONNECT"
  }
}
