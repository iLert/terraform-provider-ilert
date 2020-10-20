---
layout: "ilert"
page_title: "iLert: ilert_alert_source"
sidebar_current: "docs-ilert-resource-alert-source"
description: |-
  Creates and manages an alert source in iLert.
---

# ilert_alert_source

An [alert source](https://api.ilert.com/api-docs/#tag/Alert-Sources) represents the connection between your tools (usually a monitoring system, a ticketing tool, or an application) and iLert. We often refer to alert sources as inbound integrations.

## Example Usage

```hcl
resource "ilert_alert_source" "example" {
  name                    = "My Grafana Integration"
  integration_type        = "GRAFANA"
  escalation_policy       = ilert_escalation_policy.example.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the alert source.
- `integration_type` - (Required) The integration type of the alert source.
- `escalation_policy` - (Required) The escalation policy used by this alert source.
- `incident_creation` - (Optional) iLert receives events from your monitoring systems and can then create incidents in different ways. This option is recommended.
- `active` - (Optional) The state of the alert source. Default: true.
- `incident_priority_rule` - (Optional) The incident priority rule. This option is recommended.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the alert source.
- `status`- The status of the alert source
- `integration_key`- The integration key of the service

## Import

Services can be imported using the `id`, e.g.

```
$ terraform import ilert_alert_source.main 123456789
```
