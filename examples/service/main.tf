resource "ilert_team" "example"{
  name = "example"
}

resource "ilert_service" "example" {
  name = "example"
  status = "OPERATIONAL"
  description = "example iLert service"
  
  team {
    id = ilert_team.example.id
  }
}
