resource "ilert_service" "this" {
  name   = var.name
  status = "OPERATIONAL"

  team {
    id = ilert_team.this.id
  }
}
