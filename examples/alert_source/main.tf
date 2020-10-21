resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration from terraform"
  integration_type  = "GRAFANA"
  escalation_policy = var.escalation_policy_id
}

resource "ilert_alert_source" "example_with_support_hours" {
  name                   = "My Grafana Integration from terraform with support hours"
  integration_type       = "GRAFANA"
  escalation_policy      = var.escalation_policy_id
  incident_priority_rule = "HIGH_DURING_SUPPORT_HOURS"

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
