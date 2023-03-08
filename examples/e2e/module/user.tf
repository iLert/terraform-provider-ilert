resource "ilert_user" "this" {
  email      = "${var.name}@example.com"
  first_name = "example"
  last_name  = "example"
}
