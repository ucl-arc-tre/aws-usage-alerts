resource "aws_iam_user" "this" {
  count = local.irsa_enabled ? 0 : 1

  name = "${var.app_name}-${local.naming_infix}-user"
  path = "/"

  tags = local.aws_tags
}

moved {
  from = aws_iam_user.this
  to   = aws_iam_user.this[0]
}

resource "aws_iam_access_key" "this" {
  count = local.irsa_enabled ? 0 : 1

  user = aws_iam_user.this[0].name
}

moved {
  from = aws_iam_access_key.this
  to   = aws_iam_access_key.this[0]
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
  count = local.irsa_enabled ? 0 : 1

  name   = "${aws_iam_user.this[0].name}-policy"
  user   = aws_iam_user.this[0].name
  policy = data.aws_iam_policy_document.this.json
}

moved {
  from = aws_iam_user_policy.this
  to   = aws_iam_user_policy.this[0]
}

data "aws_eks_cluster" "this" {
  count = local.irsa_enabled ? 1 : 0
  name  = var.irsa.eks_cluster_name
}

data "aws_iam_openid_connect_provider" "eks" {
  count = local.irsa_enabled ? 1 : 0

  url = data.aws_eks_cluster.this[0].identity[0].oidc[0].issuer
}

data "aws_iam_policy_document" "irsa_assume" {
  count = local.irsa_enabled ? 1 : 0

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [data.aws_iam_openid_connect_provider.eks[0].arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_iam_openid_connect_provider.eks[0].url, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.namespace}:${kubernetes_service_account.this.metadata.0.name}"]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(data.aws_iam_openid_connect_provider.eks[0].url, "https://", "")}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "irsa" {
  count = local.irsa_enabled ? 1 : 0

  name               = "${var.app_name}-${local.naming_infix}-irsa"
  assume_role_policy = data.aws_iam_policy_document.irsa_assume[0].json
  tags               = local.aws_tags
}

resource "aws_iam_policy" "irsa" {
  count = local.irsa_enabled ? 1 : 0

  name   = "${var.app_name}-${local.naming_infix}-irsa"
  policy = data.aws_iam_policy_document.this.json
  tags   = local.aws_tags
}

resource "aws_iam_role_policy_attachment" "irsa" {
  count = local.irsa_enabled ? 1 : 0

  role       = aws_iam_role.irsa[0].name
  policy_arn = aws_iam_policy.irsa[0].arn
}
