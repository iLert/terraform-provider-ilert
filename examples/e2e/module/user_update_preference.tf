resource "ilert_user_update_preference" "this" {
  method = "EMAIL"
  type   = "ALERT_ACCEPTED"
  contact {
    id = ilert_user_email_contact.this.id
  }
  user {
    id = ilert_user.this.id
  }
}
