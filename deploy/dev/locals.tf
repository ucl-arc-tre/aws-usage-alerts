locals {
  kubeconfig_path  = "../../kubeconfig.yaml"
  k3d_cluster_name = "aws-usage-alerts"

  image = "localhost/aws-usage-alerts"

  config         = yamldecode(file("${path.module}/../../config.yaml"))
}
