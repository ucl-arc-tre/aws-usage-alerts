locals {
  health_port = 8080

  config_dir = "/etc/aws-usage-alerts"

  aws_tags = {
    Repo    = "aws-usage-alerts"
    Owner   = data.aws_caller_identity.current.arn
    AppName = var.app_name
  }
}
