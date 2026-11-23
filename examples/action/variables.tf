variable "api_token" {
  description = "ilert API token used to configure the provider"
  type        = string
}

variable "endpoint" {
  description = "ilert organization used to configure the provider"
  type        = string
  default     = "https://api.ilert.com"
}

variable "alert_source_id" {
  description = "ID of the alert source whose alerts the action posts"
  type        = string
}

variable "discord_webhook_url" {
  description = "URL of the Discord webhook the action posts to"
  type        = string
  sensitive   = true
}
