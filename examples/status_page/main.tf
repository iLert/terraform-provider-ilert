resource "ilert_service" "example" {
  name = "example"
}

resource "ilert_status_page" "example" {
  name       = "example"
  subdomain  = "example.ilert.io"
  visibility = "PUBLIC"

  service {
    id = ilert_service.example.id
  }
}

# private status page with ip whitelist enabled

# resource "ilert_status_page" "example" {
#   name         = "example"
#   subdomain    = "example.ilert.io"
#   visibility   = "PRIVATE"
#   ip_whitelist = ["###.###.###.###"]

#   service {
#     id = ilert_service.example.id
#   }
# }
