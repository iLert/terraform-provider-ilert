resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}

resource "ilert_team" "example" {
  name       = "My Team"
  visibility = "PRIVATE"

  member {
    user = ilert_user.example.id
    role = "STAKEHOLDER"
  }
}
