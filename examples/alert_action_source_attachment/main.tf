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

resource "ilert_alert_source" "bootstrap" {
  name              = "bootstrap-source"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.example.id
}

resource "ilert_alert_source" "team_a" {
  name              = "team-a-source"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.example.id
}

resource "ilert_alert_source" "team_b" {
  name              = "team-b-source"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.example.id
}

resource "ilert_connector" "example" {
  name = "shared connector"
  type = "github"

  github {
    api_key = "my api key"
  }
}

resource "ilert_alert_action" "shared" {
  name = "shared GitHub Alert Action"

  alert_source {
    id = ilert_alert_source.bootstrap.id
  }

  connector {
    id   = ilert_connector.example.id
    type = ilert_connector.example.type
  }

  github {
    owner      = "my org"
    repository = "my repo"
  }

  lifecycle {
    ignore_changes = [alert_source]
  }
}

resource "ilert_alert_action_source_attachment" "team_a_to_shared" {
  alert_action_id = ilert_alert_action.shared.id
  alert_source_id = ilert_alert_source.team_a.id
}

resource "ilert_alert_action_source_attachment" "team_b_to_shared" {
  alert_action_id = ilert_alert_action.shared.id
  alert_source_id = ilert_alert_source.team_b.id
}
