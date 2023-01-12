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

resource "ilert_status_page_group" "example" {
  name           = "example"
  status_page_id = ilert_status_page.example.id
}
