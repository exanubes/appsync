
resource "random_id" "suffix" {
  byte_length = 4
}

locals {
  name_prefix = "appsync-e2e-${random_id.suffix.hex}"
}

resource "aws_cognito_user_pool" "e2e" {
  name = local.name_prefix

  mfa_configuration = "OFF"

  username_attributes      = []
  auto_verified_attributes = []

  admin_create_user_config {
    allow_admin_create_user_only = true
  }

  password_policy {
    minimum_length    = 6
    require_lowercase = false
    require_numbers   = false
    require_symbols   = false
    require_uppercase = false
  }

  account_recovery_setting {
    recovery_mechanism {
      name     = "admin_only"
      priority = 1
    }
  }
}

resource "aws_cognito_user_pool_client" "cognito_auth" {
  name         = "${local.name_prefix}-cognito-client"
  user_pool_id = aws_cognito_user_pool.e2e.id

  generate_secret = false

  explicit_auth_flows = [
    "ALLOW_ADMIN_USER_PASSWORD_AUTH",
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
  ]

  access_token_validity = 60
  id_token_validity     = 60

  token_validity_units {
    access_token = "minutes"
    id_token     = "minutes"
  }
}

resource "aws_cognito_user_pool_domain" "e2e" {
  domain       = local.name_prefix
  user_pool_id = aws_cognito_user_pool.e2e.id
}

resource "aws_cognito_resource_server" "appsync" {
  user_pool_id = aws_cognito_user_pool.e2e.id

  identifier = local.name_prefix
  name       = local.name_prefix

  scope {
    scope_name        = "custom"
    scope_description = "Custom scope for AppSync OIDC e2e client_credentials flow"
  }
}

resource "aws_cognito_user_pool_client" "oidc_auth" {
  name         = "${local.name_prefix}-oidc-client"
  user_pool_id = aws_cognito_user_pool.e2e.id

  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows = [
    "client_credentials",
  ]

  allowed_oauth_scopes = [
    "${aws_cognito_resource_server.appsync.identifier}/custom",
  ]

  generate_secret = true

  access_token_validity = 60

  token_validity_units {
    access_token = "minutes"
  }

  depends_on = [
    aws_cognito_resource_server.appsync,
  ]
}
