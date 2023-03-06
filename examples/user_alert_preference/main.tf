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

resource "ilert_user_alert_preference" "example" {
  method = "EMAIL"
  contact {
    id = ilert_user_email_contact.example.id
  }
  delay_min = 0
  type      = "HIGH_PRIORITY"
  user {
    id = ilert_user.example.id
  }
}
