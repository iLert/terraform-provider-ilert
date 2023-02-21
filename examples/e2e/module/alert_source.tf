resource "ilert_alert_source" "this" {
  name              = var.name
  integration_type  = "GRAFANA"
  escalation_policy = ilert_escalation_policy.this.id
  team {
    id = ilert_team.this.id
  }
}
