
/*
 * Create IAM user for Serverless framework to use to deploy Bitbucket 2fa monitor
 */
module "bitbucket-2fa-monitor-serverless-user" {
  source  = "silinternational/serverless-user/aws"
  version = "0.1.0"

  app_name   = "bitbucket-2fa-mon"
  aws_region = var.aws_region
}

output "bitbucket-2fa-monitor-access-key-id" {
  value = module.bitbucket-2fa-monitor-serverless-user.aws_access_key_id
}
output "bitbucket-2fa-monitor-secret-access-key" {
  value     = module.bitbucket-2fa-monitor-serverless-user.aws_secret_access_key
  sensitive = true
}
