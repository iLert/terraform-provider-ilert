variable "organization" {
  description = "iLert organization used to configure the provider"
  type        = string
}

variable "username" {
  description = "iLert username used to configure the provider"
  type        = string
}

variable "password" {
  description = "iLert password used to configure the provider"
  type        = string
}

variable "endpoint" {
  description = "iLert organization used to configure the provider"
  type        = string
  defualt     = "https://api.ilert.com"
}
