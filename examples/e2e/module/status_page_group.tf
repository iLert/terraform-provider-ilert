resource "ilert_status_page_group" "this" {
  name = var.name
  status_page {
    id = ilert_status_page.this.id
  }
}
