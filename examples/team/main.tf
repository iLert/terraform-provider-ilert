data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_user" "example" {
  email      = "example@example.com"
  username   = "example"
  first_name = "example"
  last_name  = "example"

  high_priority_notification_preference {
    delay  = 0
    method = "EMAIL"
  }
}

resource "ilert_team" "example" {
  name = "My Team"
  # visibility = "PRIVATE"

  member {
    user = ilert_user.example.id
    # role = "STAKEHOLDER"
  }
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration"
  integration_type  = "GRAFANA"
  escalation_policy = data.ilert_escalation_policy.default.id
  team {
    id = ilert_team.example.id
  }
}
