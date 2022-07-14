data "ilert_escalation_policy" "default" {
  name = "Default"
}

data "ilert_service" "example" {
  name = "example"
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration from terraform"
  integration_type  = "GRAFANA"
  escalation_policy = data.ilert_escalation_policy.default.id
}

resource "ilert_automation_rule" "example" {
  alert_type = "CREATED"
  service_status = "OPERATIONAL"
  service {
    id = data.ilert_service.example.id
  }
   
  alert_source {
    id = ilert_alert_source.example.id
  }
}
