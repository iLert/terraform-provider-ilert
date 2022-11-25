---
layout: "ilert"
page_title: "ilert: ilert_automation_rule"
sidebar_current: "docs-ilert-resource-automation-rule"
description: |-
  Creates and manages an automation rule in ilert.
---

# ilert_automation_rule

An [automation rule](https://api.ilert.com/api-docs/#tag/Automation-Rules) is used for automatically setting the status of a service and creating incidents. They are triggered by your alert sources.

## Example Usage

```hcl
data "ilert_escalation_policy" "default" {
  name = "Default"
}

data "ilert_service" "example" {
  name = "example"
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration from terraform"
  integration_type  = "GRAFANA"
  escalation_policy = data.ilert_escalation_policy.default.id
}

resource "ilert_automation_rule" "example" {
  alert_type = "CREATED"
  service_status = "OPERATIONAL"
  service {
    id = data.ilert_service.example.id
  }
  alert_source {
    id = ilert_alert_source.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `alert_type` - (Required) The alert type of the automation rule. Allowed values are `CREATED` and `ACCEPTED`.
- `service_status` - (Required) The status in which the service is currently in. Allowed values are `OPERATIONAL`, `UNDER_MAINTENANCE`, `DEGRADED`, `PARTIAL_OUTAGE`, `MAJOR_OUTAGE`.
- `service` - (Required) The [service](#service-arguments) block.
- `alert_source` - (Required) The [alert source](#alert-source-arguments) block.
- `resolve_incident` - (Optional) Indicates whether or not the incident will be resolved automatically. Default: `false`
- `resolve_service` - (Optional) Indicates whether or not the service will be resolved automatically. Default: `true`
- `template` - (Optional) The [incident template](#incident-template-arguments) block.
- `send_notification` - (Optional) Indicates whether or not notifications will be sent. Default: `false`

#### Service Arguments

- `id` - (Required) The ID of the service.
- `name` - (Optional) The name of the service.

#### Alert Source Arguments

- `id` - (Required) The ID of the alert source.
- `name` - (Optional) The name of the alert source.

#### Incident Template Arguments

- `id` - (Required) The ID of the incident template.
- `name` - (Optional) The name of the incident template.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the automation rule.
- `alert_type` - The alert type of the automation rule.
- `service_status` - The status in which the service is currently in.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_automation_rule.main 123456789
```
