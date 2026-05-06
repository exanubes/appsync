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
      auth_type = "AMAZON_COGNITO_USER_POOLS"
      cognito_config {
        user_pool_id = aws_cognito_user_pool.dev.id
        aws_region   = data.aws_region.current.region
      }
    }

    auth_provider {
      auth_type = "AWS_LAMBDA"
      lambda_authorizer_config {
        authorizer_uri                   = aws_lambda_function.authorizer.arn
        authorizer_result_ttl_in_seconds = 300
      }
    }

    auth_provider {
      auth_type = "OPENID_CONNECT"
      openid_connect_config {
        issuer   = "https://cognito-idp.${data.aws_region.current.id}.amazonaws.com/${aws_cognito_user_pool.dev.id}"
        auth_ttl = 3600000
        iat_ttl  = 3600000
      }
    }

    connection_auth_mode {
      auth_type = "AWS_IAM"
    }

    default_publish_auth_mode {
      auth_type = "AWS_IAM"
    }

    default_subscribe_auth_mode {
      auth_type = "AWS_IAM"
    }
  }
}


resource "aws_appsync_channel_namespace" "dev" {
  name   = "appsync-dev"
  api_id = aws_appsync_api.dev.api_id

  subscribe_auth_mode {
    auth_type = "AWS_IAM"
  }

  publish_auth_mode {
    auth_type = "AWS_IAM"
  }
}

resource "aws_appsync_api_key" "dev" {
  api_id = aws_appsync_api.dev.api_id
}
