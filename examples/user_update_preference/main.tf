resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}

resource "ilert_user_email_contact" "example" {
  target = "example@example.com"
  user {
    id = ilert_user.example.id
  }
}

resource "ilert_user_update_preference" "example" {
  method = "EMAIL"
  type   = "ALERT_ACCEPTED"
  contact {
    id = ilert_user_email_contact.example.id
  }
  user {
    id = ilert_user.example.id
  }
}
