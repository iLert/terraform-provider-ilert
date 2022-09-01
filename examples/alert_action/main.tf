resource "ilert_user" "example" {
  username   = "example1"
  first_name = "example"
  last_name  = "example"
  email      = "example1@example.com"
}

resource "ilert_escalation_policy" "example" {
  name = "example"
  escalation_rule {
    escalation_timeout = 15
    user               = ilert_user.example.id
  }
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration for GitHub"
  integration_type  = "GRAFANA"
  escalation_policy = ilert_escalation_policy.default.id
}

resource "ilert_connector" "example" {
  name = "My GitHub Connector"
  type = "github"

  github {
    api_key = "my api key"
  }
}

resource "ilert_alert_action" "example" {
  name = "My GitHub Alert Action"

  alert_source {
    id = ilert_alert_source.example.id
  }

  connector {
    id   = ilert_connector.example.id
    type = ilert_connector.example.type
  }

  github {
    owner      = "my org"
    repository = "my repo"
  }
}
