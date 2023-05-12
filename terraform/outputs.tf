
output "bitbucket-2fa-monitor-access-key-id" {
  value = module.bitbucket-2fa-monitor-serverless-user.aws_access_key_id
}

output "bitbucket-2fa-monitor-secret-access-key" {
  value     = module.bitbucket-2fa-monitor-serverless-user.aws_secret_access_key
  sensitive = true
}
