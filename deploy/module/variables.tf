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

variable "unique_infix" {
  type    = string
  default = ""
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

variable "update_delay_seconds" {
  type        = number
  description = "Number of seconds between updates"
  default     = 60

  validation {
    condition     = var.update_delay_seconds >= 1
    error_message = "Delay must be at least one second"
  }
}

variable "config" {
  type = object({
    groupTagKey    = string
    storageBackend = string
    adminEmails    = list(string)
    groups = map(object({
      threshold = number
    }))
  })
  description = "Configuration map"

  validation {
    condition     = contains(["inMemory", "configMap"], var.config.storageBackend)
    error_message = "storageBackend must be one of: {inMemory, configMap}"
  }
}
