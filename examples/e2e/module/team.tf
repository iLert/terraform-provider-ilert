resource "ilert_team" "this"{
  name = var.name

  member {
    user = ilert_user.this.id
    role = "ADMIN"
  }
}
