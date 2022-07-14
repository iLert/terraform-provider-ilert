data "ilert_user" "example" {
  email = "example@example.com"
}

data "ilert_schedule" "example" {
  name = "example"
}

resource "ilert_escalation_policy" "example" {
  name = "example"

  escalation_rule {
    escalation_timeout = 5
    schedule           = data.ilert_schedule.example.id
  }

  escalation_rule {
    escalation_timeout = 15
    user               = data.ilert_user.example.id
  }
}
