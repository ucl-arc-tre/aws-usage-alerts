resource "aws_sns_topic" "this" {
  name = "${var.app_name}-${local.naming_infix}-topic"
}

resource "aws_sns_topic_subscription" "main" {
  for_each = toset(var.config.adminEmails)

  topic_arn = aws_sns_topic.this.arn
  protocol  = "email"

  endpoint = each.value
}
