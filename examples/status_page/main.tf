data "ilert_service" "example" {
  name = "example"
}

resource "ilert_status_page" "example" {
  name = "example"
  service = data.ilert_service.example.id
}
