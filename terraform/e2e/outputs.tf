output "user_pool_id" {
  value = aws_cognito_user_pool.e2e.id
}

output "cognito_user_pool_client_id" {
  value = aws_cognito_user_pool_client.cognito_auth.id
}

output "cognito_domain" {
  value = "https://${aws_cognito_user_pool_domain.e2e.domain}.auth.${var.aws_region}.amazoncognito.com"
}

output "oidc_user_pool_client_id" {
  value = aws_cognito_user_pool_client.oidc_auth.id
}

output "oidc_user_pool_client_secret" {
  value     = aws_cognito_user_pool_client.oidc_auth.client_secret
  sensitive = true
}

output "oidc_scope" {
  value = "${aws_cognito_resource_server.appsync.identifier}/custom"
}

output "cognito_issuer" {
  value = "https://cognito-idp.${var.aws_region}.amazonaws.com/${aws_cognito_user_pool.e2e.id}"
}

output "appsync_http_endpoint" {
  value = "https://${aws_appsync_api.e2e.dns.HTTP}/event"
}

output "appsync_ws_endpoint" {
  value = "wss://${aws_appsync_api.e2e.dns.REALTIME}/event/realtime"
}

output "appsync_api_key" {
  value     = aws_appsync_api_key.e2e.key
  sensitive = true
}
