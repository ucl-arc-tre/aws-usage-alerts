variable "namespace" {
  type        = string
  description = "Name of the namespace to create"
  default     = "aws-usage-alerts"
}

variable "image" {
  type        = string
  description = "Full tag of the image to use e.g. `ghcr.io/ucl-arc-tre/aws-usage-alerts:latest`"
}

variable "app_name" {
  type    = string
  default = "aws-usage-alerts"
}

variable "replicas" {
  type        = number
  default     = 1
  description = "Number of replicas of the app to deploy"
}

variable "debug_logging" {
  type        = bool
  description = "Should debug logging be enabled?"
  default     = false
}

variable "trace_logging" {
  type        = bool
  description = "Should trace logging be enabled?"
  default     = false
}

variable "config_file_content" {
  type        = string
  description = "File contents of the config.yaml file"
}

variable "email_addresses" {
	type = list(string)
	description = "List of email addresses to notify"
}
