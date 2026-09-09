---
layout: "ilert"
page_title: "ilert: ilert_service"
sidebar_current: "docs-ilert-resource-service"
description: |-
  Creates and manages a service in ilert.
---

# ilert_service

A [service](https://api.ilert.com/api-docs/#tag/Services) serves as a starting point when manually creating incidents. It is also used in alert source automation rules for automatically creating incidents.

## Example Usage

```hcl
resource "ilert_service" "example" {
  name = "example"
  status = "OPERATIONAL"
  description = "example ilert service"

  labels = {
    env = "production"
  }

  link {
    href = "https://example.com/runbook"
    text = "Runbook"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the service.
- `alias` - (Optional) The alias of the service.
- `status` - (Optional) The status of the service. Allowed values are `OPERATIONAL`, `UNDER_MAINTENANCE`, `DEGRADED`, `PARTIAL_OUTAGE`, `MAJOR_OUTAGE`.
- `description` - (Optional) The description of the service.
- `one_open_incident_only` - (Optional) Indicates whether or not only one incident should be opened. Default: `false`
- `show_uptime_history` - (Optional) Indicates whether or not the uptime history should be shown. Default: `true`
- `labels` - (Optional) A map of free-form key-value labels assigned to the service. The labels are managed by Terraform only once the attribute is declared: removing it after it was declared clears the labels on the server, while a configuration that never declared it leaves labels assigned in the web app untouched.
- `icon_url` - (Optional) The URL of an icon shown for the service. Like `link` and unlike `labels`, this is owned by Terraform outright: the API clears it whenever the field is absent from an update, so an icon set in the web app on a service managed by Terraform is removed on the next apply.
- `link` - (Optional) One or more [link](#link-arguments) blocks displayed on the service. Unlike `team`, the order in which the blocks are declared is significant: the API preserves it and displays the links in it. Links do **not** follow the `labels` rule: the API clears them whenever the field is absent from an update, so Terraform always sends them and owns them outright. A link added in the web app on a service managed by Terraform is removed on the next apply, and appears in the plan as a removal beforehand.
- `team` - (Optional) One or more [team](#team-arguments) blocks. The order in which the blocks are declared is not significant.

#### Link Arguments

- `href` - (Required) The URL the link points to.
- `text` - (Optional) The display text of the link.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the service.
- `name` - The name of the service.
- `public_status` - The status shown for the service on status pages, derived by the API from `status`.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_service.main 123456789
```
