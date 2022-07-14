data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_uptime_monitor" "terraform" {
  name                                = "terraform.io"
  region                              = "EU"
  escalation_policy                   = data.ilert_escalation_policy.default.id
  interval_sec                        = 900
  timeout_ms                          = 10000
  create_incident_after_failed_checks = 2
  check_type                          = "http"

  check_params {
    url = "https://www.terraform.io"
  }
}
