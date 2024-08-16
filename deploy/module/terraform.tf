terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">=5.51.1"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = ">=2.30.0"
    }
  }
}
