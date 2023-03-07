resource "ilert_user" "example" {
  first_name = "example"
  last_name  = "example"
  email      = "example@example.com"
}

resource "ilert_schedule" "example" {
  name     = "example"
  timezone = "Europe/Berlin"
  type     = "STATIC"
  shift {
    user  = ilert_user.example.id
    start = "2023-09-01T08:00"
    end   = "2023-09-02T08:00"
  }
}

resource "ilert_escalation_policy" "example" {
  name = "example"

  escalation_rule {
    escalation_timeout = 15
    users {
      id = ilert_user.example.id
    }
    schedules {
      id = ilert_schedule.example.id
    }
  }
}
