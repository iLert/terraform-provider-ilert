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

resource "ilert_user_update_preference" "example" {
  method = "EMAIL"
  type   = "ALERT_ACCEPTED"
  contact {
    id = data.ilert_user_email_contact.example.id
  }
  user {
    id = ilert_user.example.id
  }
}
