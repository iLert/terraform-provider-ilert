resource "ilert_user_duty_preference" "this" {
  method = "EMAIL"
  contact {
    id = data.ilert_user_email_contact.this.id
  }
  before_min = 0
  type       = "ON_CALL"
  user {
    id = ilert_user.this.id
  }
}
