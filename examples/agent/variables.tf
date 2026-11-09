variable "api_token" {
  description = "ilert API token used to configure the provider"
  type        = string
}

variable "endpoint" {
  description = "ilert organization used to configure the provider"
  type        = string
  default     = "https://api.ilert.com"
}

variable "github_token" {
  description = "GitHub token the GitHub MCP server of the agent authenticates with"
  type        = string
  sensitive   = true
}
