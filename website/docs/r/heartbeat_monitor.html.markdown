---
layout: "ilert"
page_title: "ilert: ilert_heartbeat_monitor"
sidebar_current: "docs-ilert-resource-heartbeat-monitor"
description: |-
    Creates and manages a heartbeat monitor in ilert.
---

# ilert_heartbeat_monitor

A [heartbeat monitor](https://api.ilert.com/api-docs/#tag/heartbeat-monitors) allows you to monitor services, devices, or workflows via receiving signals in intervals, creating alerts through an alert source in ilert if the interval is not met.

## Example Usage

```hcl
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

resource "ilert_alert_source" "example" {
  name              = "My API integration from terraform"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.example.id
}

resource "ilert_heartbeat_monitor" "example" {
  name          = "example"
  interval_sec  = 60
  alert_summary = "Heartbeat monitor alert"
  alert_source {
    id = ilert_alert_source.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the heartbeat monitor.
- `interval_sec` - (Required) The interval in seconds of the heartbeat monitor. Minimum value: 25.
- `alert_summary` - (Optional) The summary of the heartbeat monitor alert.
- `alert_source` - (Optional) One [alert-source](#alert-source-arguments) block.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `state` - (Computed) The state of the heartbeat monitor.
- `created_at` - (Computed) The creation date of the heartbeat monitor.
- `updated_at` - (Computed) The latest date the heartbeat monitor was updated at.
- `integration_key` - (Computed) The integration key of the heartbeat monitor.
- `integration_url` - (Computed) The integration url of the heartbeat monitor.

#### Alert Source Arguments

- `id` - (Required) The ID of the alert source.
- `name` - (Optional) The name of the alert source.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Import

heartbeat monitors can be imported using the `id`, e.g.

```sh
$ terraform import ilert_heartbeat_monitor.main 123456789
```
