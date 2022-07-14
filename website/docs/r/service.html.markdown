---
layout: "ilert"
page_title: "iLert: ilert_service"
sidebar_current: "docs-ilert-resource-service"
description: |-
  Creates and manages a service in iLert.
---

# ilert_service

A [service](https://api.ilert.com/api-docs/#tag/Services) serves as a starting point when manually creating incidents. It is also used in alert source automation rules for automatically creating incidents.

## Example Usage

```hcl
resource "ilert_team" "example"{
  name = "example"
}

resource "ilert_service" "example" {
  name = "example"
  status = "OPERATIONAL"
  description = "example iLert service"
  team {
    id = ilert_team.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the service.
- `status` - (Optional) The status of the service. Allowed values are `OPERATIONAL`, `UNDER_MAINTENANCE`, `DEGRADED`, `PARTIAL_OUTAGE`, `MAJOR_OUTAGE`.
- `description` - (Optional) The description of the service.
- `one_open_incident_only` - (Optional) Indicates whether or not only one incident should be opened. Default: `false`
- `show_uptime_history` - (Optional) Indicates whether or not the uptime history should be shown. Default: `true`
- `team` - (Optional) One or more [team](#team-arguments) blocks.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the service.
- `name` - The name of the service.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_service.main 123456789
```
