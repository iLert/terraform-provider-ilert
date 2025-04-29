resource "ilert_user" "example" {
  first_name = "example"
  last_name  = "example"
  email      = "example@example.com"
}

resource "ilert_escalation_policy" "example" {
  name = "example"
  escalation_rule {
    escalation_timeout = 15
    user               = ilert_user.example.id
  }
}

resource "ilert_alert_source" "example" {
  name              = "My API integration from terraform"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.example.id
}

resource "ilert_heartbeat_monitor" "example" {
  name          = "example"
  interval_sec  = 60
  alert_summary = "Heartbeat monitor alert"
  alert_source {
    id = ilert_alert_source.example.id
  }
}
