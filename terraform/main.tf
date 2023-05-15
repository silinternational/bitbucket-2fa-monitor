
/*
 * Create IAM user for Serverless framework to use to deploy Bitbucket 2fa monitor
 */
module "bitbucket-2fa-monitor-serverless-user" {
  source  = "silinternational/serverless-user/aws"
  version = "0.1.3"

  app_name   = "bitbucket-2fa-mon"
  aws_region = var.aws_region
}
