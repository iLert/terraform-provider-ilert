resource "ilert_user_subscription_preference" "this" {
  method = "EMAIL"
  contact {
    id = data.ilert_user_email_contact.this.id
  }
  user {
    id = ilert_user.this.id
  }
}
