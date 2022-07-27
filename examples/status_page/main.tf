resource "ilert_service" "example" {
  name = "example"
}

resource "ilert_status_page" "example" {
  name       = "example"
  subdomain  = "example.ilerthq.com"
  visibility = "PUBLIC"

  service {
    id = ilert_service.example.id
  }
}
