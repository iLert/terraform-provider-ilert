terraform {
  required_providers {
    ilert = {
      source  = "iLert/ilert"
      version = "~> 1.0"
    }
  }
}

provider "ilert" {
  endpoint     = var.endpoint
  organization = var.organization
  username     = var.username
  password     = var.password
}
