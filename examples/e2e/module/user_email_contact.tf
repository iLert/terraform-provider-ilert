resource "ilert_user_email_contact" "this" {
  target = "${var.name}@fake.com"
  user {
    id = ilert_user.this.id
  }
}
