terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.51.1"
    }

    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "2.30.0"
    }
  }
}

resource "kubernetes_namespace" "this" {
  metadata {
    name = var.namespace
  }
}

resource "kubernetes_network_policy" "deny_all" {
  metadata {
    name      = "deny-all"
    namespace = kubernetes_namespace.this.metadata.0.name
  }

  spec {
    pod_selector {}
    policy_types = ["Ingress"]
  }
}

resource "kubernetes_deployment" "this" {
  metadata {
    name      = "aws-usage-alerts"
    namespace = kubernetes_namespace.this.metadata.0.name
    labels = {
      "app.kubernetes.io/name" = var.app_name
    }
  }

  spec {
    replicas = 1 # WARNING: non-leased configmap backends do not support >1 replica

    strategy {
      # Wait for app operations to finish, rather than running rolling updates, as
      # mulitple running pods may result in race conditions with state updates
      type = "Recreate"
    }

    selector {
      match_labels = {
        "app.kubernetes.io/name" = var.app_name
      }
    }

    template {
      metadata {
        labels = {
          "app.kubernetes.io/name" = var.app_name
        }
      }

      spec {
        restart_policy       = "Always"
        service_account_name = kubernetes_service_account.this.metadata.0.name

        container {
          image             = var.image
          name              = "app"
          image_pull_policy = "IfNotPresent"

          env {
            name  = "TRACE"
            value = var.trace_logging ? "true" : "false"
          }

          env {
            name  = "DEBUG"
            value = var.debug_logging ? "true" : "false"
          }

          env {
            name  = "HEALTH_PORT"
            value = local.health_port
          }

          env {
            name  = "CONFIG_DIR"
            value = local.config_dir
          }

          env {
            name  = "AWS_REGION"
            value = data.aws_region.current.name
          }

          env {
            name  = "SNS_TOPIC_ARN"
            value = aws_sns_topic.this.arn
          }

          env {
            name  = "UPDATE_DELAY_SECONDS"
            value = var.update_delay_seconds
          }

          env {
            name = "AWS_ACCESS_KEY_ID"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.aws_keys.metadata.0.name
                key  = "AWS_ACCESS_KEY_ID"
              }
            }
          }

          env {
            name = "AWS_SECRET_ACCESS_KEY"
            value_from {
              secret_key_ref {
                name = kubernetes_secret.aws_keys.metadata.0.name
                key  = "AWS_SECRET_ACCESS_KEY"
              }
            }
          }

          volume_mount {
            name       = kubernetes_config_map.config.metadata.0.name
            read_only  = true
            mount_path = local.config_dir
          }

          liveness_probe {
            http_get {
              path = "/ping"
              port = local.health_port
            }
          }

          resources {
            requests = {
              cpu    = "100m"
              memory = "128Mi"
            }
            limits = {
              cpu    = "500m"
              memory = "512Mi"
            }
          }

          security_context {
            run_as_user                = 1000
            run_as_group               = 1000
            run_as_non_root            = true
            allow_privilege_escalation = false
            read_only_root_filesystem  = true
            capabilities {
              drop = ["ALL"]
            }
          }
        }

        volume {
          name = kubernetes_config_map.config.metadata.0.name
          config_map {
            name = kubernetes_config_map.config.metadata.0.name
          }
        }

        security_context {
          run_as_user     = 1000
          run_as_non_root = true
          seccomp_profile {
            type = "RuntimeDefault"
          }
        }
      }
    }
  }
}

resource "kubernetes_config_map" "config" {
  metadata {
    name      = "config"
    namespace = kubernetes_namespace.this.metadata.0.name
  }

  data = {
    "config.yaml" = yamlencode(var.config)
  }
}

resource "kubernetes_secret" "aws_keys" {
  metadata {
    name      = "aws-keys"
    namespace = kubernetes_namespace.this.metadata.0.name
  }

  data = {
    "AWS_ACCESS_KEY_ID"     = aws_iam_access_key.this.id
    "AWS_SECRET_ACCESS_KEY" = aws_iam_access_key.this.secret
  }
}


resource "kubernetes_service_account" "this" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.this.metadata.0.name
  }
}

resource "kubernetes_role" "this" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.this.metadata.0.name
  }

  # not needed with in-memory
  rule {
    api_groups = [""]
    resources  = ["configmaps"]
    verbs      = ["get", "list", "create", "update"]
  }
}

resource "kubernetes_role_binding" "this" {
  metadata {
    name      = var.app_name
    namespace = kubernetes_namespace.this.metadata.0.name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.this.metadata.0.name
    namespace = kubernetes_service_account.this.metadata.0.namespace
  }

  role_ref {
    kind      = "Role"
    name      = kubernetes_role.this.metadata.0.name
    api_group = "rbac.authorization.k8s.io"
  }
}

resource "aws_sns_topic" "this" {
  name = "${var.app_name}-topic"
}

resource "aws_sns_topic_subscription" "main" {
  for_each = toset(var.config.adminEmails)

  topic_arn = aws_sns_topic.this.arn
  protocol  = "email"

  endpoint = each.value
}
