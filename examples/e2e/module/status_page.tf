resource "ilert_status_page" "this" {
  name       = var.name
  subdomain  = "${var.name}.ilert.io"
  visibility = "PUBLIC"

  service {
    id = ilert_service.this.id
  }
}
