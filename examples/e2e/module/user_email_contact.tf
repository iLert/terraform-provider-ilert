data "ilert_user_email_contact" "this" {
  target = "${var.name}@example.com"
  user {
    id = ilert_user.this.id
  }
}

resource "ilert_user_email_contact" "this_new" {
  target = "${var.name}_new@example.com"
  user {
    id = ilert_user.this.id
  }
}
