output "HTTP_ENDPOINT" {
  value = format("%s://%s/%s", "https", aws_appsync_api.dev.dns.HTTP, "event")
}

output "WS_ENDPOINT" {
  value = format("%s://%s/%s", "wss", aws_appsync_api.dev.dns.REALTIME, "event/realtime")
}

output "WS_CHANNEL" {
  value = aws_appsync_channel_namespace.dev.name
}

output "REGION" {
  value = data.aws_region.current.id
}
