data "aws_region" "current" {}

resource "aws_appsync_api" "dev" {
  name = "appsync-dev"

  event_config {
    auth_provider {
      auth_type = "AWS_IAM"
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

