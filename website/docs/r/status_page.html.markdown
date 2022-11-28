---
layout: "ilert"
page_title: "ilert: ilert_status_page"
sidebar_current: "docs-ilert-resource-status-page"
description: |-
  Creates and manages a status page in ilert.
---

# ilert_status_page

A [status page](https://api.ilert.com/api-docs/#tag/Status-Pages) is connected with your monitoring tools and let you update your status page automatically with automation rules or manually with a single click.

## Example Usage

```hcl
data "ilert_service" "example" {
  name = "example"
}

resource "ilert_status_page" "example" {
  name       = "example"
  subdomain  = "example.ilerthq.com"
  visibility = "PUBLIC"

  service {
    id = data.ilert_service.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the status page.
- `subdomain` - (Required) The ilert domain of the status page. Format: `[your status page].ilerthq.com`
- `visibility` - (Required) The visibility of the status page. Allowed values are `PUBLIC` and `PRIVATE`.
- `service` - (Required) One or more [service](#service-arguments) blocks.
- `domain` - (Optional) The custom domain of the status page.
- `timezone` - (Optional) The timezone of the status page. In timezone format, e.g. `Europe/Berlin`, `America/New_York`, `America/Los_Angeles`, `Asia/Istanbul`.
- `custom_css` - (Optional) Custom CSS Styles for the status page.
- `favicon_url` - (Optional) The favicon of the status page.
- `logo_url` - (Optional) The logo of the status page.
- `hidden_from_search` - (Optional) Indicates whether or not the status page is hidden from search.
- `show_subscribe_action` - (Optional) Indicates whether or not the status page is hidden from search.
- `show_incident_history_option` - (Optional) Indicates whether or not the incident history option should be shown.
- `page_title` - (Optional) The title of the status page.
- `page_description` - (Optional) The description of the status page.
- `logo_redirect_url` - (Optional) The redirect url for the status page logo.
- `team` - (Optional) One or more [team](#team-arguments) blocks.

#### Service Arguments

- `id` - (Required) The ID of the service.
- `name` - (Optional) The name of the service.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the status page.
- `name` - The name of the status page.
- `subdomain` - The ilert domain of the status page.
- `visibility` - The visibility of the status page.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_status_page.main 123456789
```
