
resource "aws_cognito_user_pool" "dev" {
  name = "appsync-dev"

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

resource "aws_cognito_user_pool_client" "dev" {
  name            = "appsync-dev-client"
  user_pool_id    = aws_cognito_user_pool.dev.id
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

resource "aws_cognito_user_pool_domain" "dev" {
  domain       = "appsync-dev-exanubes"
  user_pool_id = aws_cognito_user_pool.dev.id
}

resource "aws_cognito_resource_server" "appsync" {
  user_pool_id = aws_cognito_user_pool.dev.id

  identifier = "exanubes"
  name       = "exanubes"

  scope {
    scope_name        = "custom"
    scope_description = "Need a custom scope for making oidc client work with client_credentials oauth flow"
  }
}

resource "aws_cognito_user_pool_client" "oidc" {
  name         = "appsync-dev-oidc-client"
  user_pool_id = aws_cognito_user_pool.dev.id

  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["client_credentials"]
  allowed_oauth_scopes = [
    "exanubes/custom"
  ]
  generate_secret = true
  depends_on      = [aws_cognito_resource_server.appsync]
}
