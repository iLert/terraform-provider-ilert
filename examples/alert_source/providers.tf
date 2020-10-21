terraform {
  required_providers {
    ilert = {
      source  = "iLert/ilert"
      version = "0.1.5"
    }
  }
}

provider "ilert" {
  organization = var.organization
  username     = var.username
  password     = var.password
}
