terraform {
  required_providers {
    ilert = {
      source  = "iLert/ilert"
      version = "~> 2.0"
    }
  }
}

provider "ilert" {
  endpoint             = var.endpoint
  api_token            = var.api_token
  early_access_enabled = ["ilert_agent", "ilert_credential"]
}
