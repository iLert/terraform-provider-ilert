resource "ilert_escalation_policy" "this" {
  name = var.name

  escalation_rule {
    escalation_timeout = 15
    user               = ilert_user.this.id
  }

  team {
    id = ilert_team.this.id
  }
}
