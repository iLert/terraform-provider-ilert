resource "ilert_service" "example" {
  name = "example"
}

resource "ilert_status_page" "example_public" {
  name       = "example_public"
  subdomain  = "example-public.ilert.io"
  visibility = "PUBLIC"

  service {
    id = ilert_service.example.id
  }
}

# private status page with ip whitelist enabled

resource "ilert_status_page" "example_private" {
  name         = "example_private"
  subdomain    = "example-private.ilert.io"
  visibility   = "PRIVATE"
  ip_whitelist = ["###.###.###.###"]

  service {
    id = ilert_service.example.id
  }
}
