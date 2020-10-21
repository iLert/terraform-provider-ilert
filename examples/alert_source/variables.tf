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

variable "escalation_policy_id" {
  description = "iLert escalation policy id used for alert source"
  type        = number
}
