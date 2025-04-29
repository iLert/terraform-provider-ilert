resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
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
  integration_type  = "EMAIL2"
  email             = "example@ 'your tenant' .ilert.eu"
  escalation_policy = ilert_escalation_policy.example.id

  alert_creation = "OPEN_RESOLVE_ON_EXTRACTION"
  alert_key_template {
    text_template = "{{ subject.splitTakeAt(\"my server\", 0) }}"
  }

  event_type_filter_create  = "(event.customDetails.body contains_any [\"alarm\"])"
  event_type_filter_resolve = "(event.customDetails.body in [\"resolve\"])"
}
