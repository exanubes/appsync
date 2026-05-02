output "HTTP_ENDPOINT" {
  value = format("%s://%s/%s", "https", aws_appsync_api.dev.dns.HTTP, "event")
}

output "WS_ENDPOINT" {
  value = format("%s://%s/%s", "wss", aws_appsync_api.dev.dns.REALTIME, "event/realtime")
}

output "CHANNEL" {
  value = aws_appsync_channel_namespace.dev.name
}

output "APPSYNC_API_KEY" {
  value     = aws_appsync_api_key.dev.key
  sensitive = true
}

output "AWS_REGION" {
  value = data.aws_region.current.id
}

output "USER_POOL" {
  value = aws_cognito_user_pool.dev.id
}

output "USER_POOL_CLIENT" {
  value = aws_cognito_user_pool_client.dev.id
}
