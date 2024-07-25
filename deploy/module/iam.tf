resource "aws_iam_user" "this" {
  name = "${var.app_name}-user"
  path = "/"

  tags = local.aws_tags
}

resource "aws_iam_access_key" "this" {
  user = aws_iam_user.this.name
}

data "aws_iam_policy_document" "this" {
  statement {
    effect = "Allow"
    actions = [
      "elasticfilesystem:DescribeFileSystems",
      "ec2:DescribeInstances",
      "pricing:GetProducts",
      "pricing:GetAttributeValues",
      "pricing:DescribeServices"
    ]
    resources = ["*"] # todo
  }

  statement {
    effect = "Allow"
    actions = [
      "SNS:Publish"
    ]
    resources = [aws_sns_topic.this.arn]
  }
}

resource "aws_iam_user_policy" "this" {
  name   = "${aws_iam_user.this.name}-policy"
  user   = aws_iam_user.this.name
  policy = data.aws_iam_policy_document.this.json
}
