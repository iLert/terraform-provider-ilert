terraform {
  required_providers {
    ilert = {
      source  = "iLert/ilert"
      version = "~> 1.7"
    }
  }
}

provider "ilert" {
  endpoint  = var.endpoint
  api_token = var.api_token
}
