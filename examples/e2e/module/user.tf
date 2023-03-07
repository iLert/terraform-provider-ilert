resource "ilert_user" "this" {
  email      = "${var.name}@fake.com"
  first_name = "fake"
  last_name  = "fake"
}
