resource "ilert_user" "example" {
  email      = "example@example.com"
  username   = "example"
  first_name = "example"
  last_name  = "example"

  mobile {
    region_code = "DE"
    number      = "+491758250853"
  }

  high_priority_notification_preference {
    method = "EMAIL"
    delay  = 0
  }

  low_priority_notification_preference {
    method = "EMAIL"
    delay  = 0
  }

  on_call_notification_preference {
    method     = "EMAIL"
    before_min = 60
  }
}

resource "ilert_escalation_policy" "example" {
  name = "example"

  escalation_rule {
    escalation_timeout = 15
    user               = ilert_user.example.id
  }
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration from terraform"
  integration_type  = "GRAFANA"
  escalation_policy = ilert_escalation_policy.example.id
}

resource "ilert_alert_source" "example_with_support_hours" {
  name                = "My Grafana Integration from terraform with support hours"
  integration_type    = "GRAFANA"
  escalation_policy   = ilert_escalation_policy.example.id
  alert_priority_rule = "HIGH_DURING_SUPPORT_HOURS"

  support_hours {
    timezone = "Europe/Berlin"

    support_days {
      monday {
        start = "08:00"
        end   = "17:00"
      }

      tuesday {
        start = "08:00"
        end   = "17:00"
      }

      wednesday {
        start = "08:00"
        end   = "17:00"
      }

      thursday {
        start = "08:00"
        end   = "17:00"
      }

      friday {
        start = "08:00"
        end   = "17:00"
      }
    }
  }
}

resource "ilert_alert_source" "example_email" {
  name              = "My Email Integration from terraform"
  integration_type  = "EMAIL"
  email             = "support2@yacut.ilertnow.com"
  escalation_policy = ilert_escalation_policy.example.id

  alert_creation = "OPEN_RESOLVE_ON_EXTRACTION"
  resolve_key_extractor {
    field    = "EMAIL_SUBJECT"
    criteria = "ALL_TEXT_BEFORE"
    value    = "my server"
  }

  email_filtered = true
  email_predicate {
    field    = "EMAIL_BODY"
    criteria = "CONTAINS_STRING"
    value    = "alarm"
  }

  email_resolve_filtered = true
  email_resolve_predicate {
    field    = "EMAIL_BODY"
    criteria = "CONTAINS_STRING"
    value    = "resolve"
  }

  team {
    id = 0000
  }

  # @ deprecated
  # team = [0000]
}
