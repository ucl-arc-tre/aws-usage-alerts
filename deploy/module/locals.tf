locals {
  health_port = 8080

  config_dir = "/etc/aws-usage-alerts"

  naming_infix = var.unique_infix != "" ? var.unique_infix : random_string.infix.result

  irsa_enabled = var.irsa != null

  aws_tags = {
    Repo    = "aws-usage-alerts"
    Owner   = data.aws_caller_identity.current.arn
    AppName = var.app_name
  }
}
