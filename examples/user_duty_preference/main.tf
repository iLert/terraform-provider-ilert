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

resource "ilert_user_duty_preference" "example" {
  method = "EMAIL"
  contact {
    id = data.ilert_user_email_contact.example.id
  }
  before_min = 0
  type       = "ON_CALL"
  user {
    id = ilert_user.example.id
  }
}
