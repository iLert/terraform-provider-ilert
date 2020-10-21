variable "organization" {
  description = "iLert organization used to configure the provider"
  type        = string
  default     = "yacut"
}

variable "username" {
  description = "iLert username used to configure the provider"
  type        = string
  default     = "yacut"
}

variable "password" {
  description = "iLert password used to configure the provider"
  type        = string
  default     = "kZjgEKL4guyCTQY"
}

variable "escalation_policy_id" {
  description = "iLert escalation policy id used for alert source"
  type        = number
  default     = 2195678
}
