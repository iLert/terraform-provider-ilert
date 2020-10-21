terraform {
  required_providers {
    ilert = {
      source  = "iLert/ilert"
      version = "0.2.0"
    }
  }
}

provider "ilert" {
  organization = var.organization
  username     = var.username
  password     = var.password
}
