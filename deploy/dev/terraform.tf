terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.30.0"
    }
  }
}

provider "kubernetes" {
  config_path = local.kubeconfig_path
}

provider "aws" {
  region = "eu-west-2"
}
