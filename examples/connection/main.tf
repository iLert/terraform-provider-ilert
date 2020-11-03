data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration for GitHub 1"
  integration_type  = "GRAFANA"
  escalation_policy = data.ilert_escalation_policy.default.id
}

resource "ilert_connector" "example" {
  name = "My GitHub Connector"
  type = "github"

  github {
    api_key = "my api key"
  }
}

resource "ilert_connection" "example" {
  name = "My GitHub Connection"

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
