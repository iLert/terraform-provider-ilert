resource "ilert_user" "example" {
  email              = "example@example.com"
  first_name         = "example"
  last_name          = "example"
  send_no_invitation = true
}
