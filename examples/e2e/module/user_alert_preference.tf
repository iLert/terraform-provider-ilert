resource "ilert_user_alert_preference" "this" {
  method = "EMAIL"
  contact {
    id = ilert_user_email_contact.this.id
  }
  delay_min = 0
  type      = "HIGH_PRIORITY"
  user {
    id = ilert_user.this.id
  }
}
