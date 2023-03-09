resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}

data "ilert_user_email_contact" "example" {
  target = "example@example.com"
  user {
    id = ilert_user.example.id
  }
}

resource "ilert_user_email_contact" "example_new" {
  target = "example_new@example.com"
  user {
    id = ilert_user.example.id
  }
}
