variable "api_token" {
  description = "iLert API token used to configure the provider"
  type        = string
}

variable "endpoint" {
  description = "iLert organization used to configure the provider"
  type        = string
  default     = "https://api.ilert.com"
}
