---
layout: "ilert"
page_title: "ilert: ilert_incident_template"
sidebar_current: "docs-ilert-resource-incident-template"
description: |-
  Creates and manages an incident template in ilert.
---

# ilert_incident_template

An [incident template](https://api.ilert.com/api-docs/#tag/Incident-Templates) serves as a starting point when manually creating incidents. It is also used in alert source automation rules for automatically creating incidents.

## Example Usage

```hcl
resource "ilert_incident_template" "example" {
  name              = "example"
  status            = "INVESTIGATING"
  send_notification = true
  summary           = "example_incident_template"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the incident template.
- `status` - (Required) The status in which the incident should be in. Allowed values are `INVESTIGATING`, `IDENTIFIED`, `MONITORING`, `RESOLVED`.
- `summary` - (Required) The summary of the incident.
- `message` - (Optional) The message of the incident.
- `send_notification` - (Optional) Indicates whether or not notifications should be sent.
- `team` - (Optional) One or more [team](#team-arguments) blocks.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the incident template.
- `name` - The name of the incident template.
- `status` - The status in which the incident should be in.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_incident_template.main 123456789
```
