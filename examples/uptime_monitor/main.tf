resource "ilert_user" "example" {
  first_name = "example"
  last_name  = "example"
  email      = "example@example.com"
}

resource "ilert_escalation_policy" "example" {
  name = "example"

  escalation_rule {
    escalation_timeout = 15
    users {
      id = ilert_user.example.id
    }
  }
}

resource "ilert_uptime_monitor" "terraform" {
  name                                = "terraform.io"
  region                              = "EU"
  escalation_policy                   = ilert_escalation_policy.example.id
  interval_sec                        = 900
  timeout_ms                          = 10000
  create_incident_after_failed_checks = 2
  check_type                          = "http"

  check_params {
    url = "https://www.terraform.io"
  }
}
