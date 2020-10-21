resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration from terraform"
  integration_type  = "GRAFANA"
  escalation_policy = var.escalation_policy_id
}
