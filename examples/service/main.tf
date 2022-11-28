resource "ilert_team" "example" {
  name = "example"
}

resource "ilert_service" "example" {
  name        = "example"
  status      = "OPERATIONAL"
  description = "example ilert service"

  team {
    id = ilert_team.example.id
  }
}
