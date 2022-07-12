data "ilert_team" "example"{
  name = "example"
}

resource "ilert_team" "example" {
  id = data.ilert_team.example.id
}

resource "ilert_service" "example" {
  name = "example"
  status = "OPERATIONAL"
  description = "example iLert service"
  teams = [ilert_team.example.id]
}
