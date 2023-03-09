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
    start = formatdate("YYYY-MM-DD'T'hh:mm:ss", timeadd(timestamp(), "730h"))
    end   = formatdate("YYYY-MM-DD'T'hh:mm:ss", timeadd(timestamp(), "754h"))
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
