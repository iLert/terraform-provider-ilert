terraform {
  required_providers {
    ilert = {
      source  = "iLert/ilert"
      version = "~> 1.0"
    }
  }
}

provider "ilert" {
  endpoint     = "https://api.ilert.dev"
  organization = var.organization
  username     = var.username
  password     = var.password
}
