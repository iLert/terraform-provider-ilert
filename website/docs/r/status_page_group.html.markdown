---
layout: "ilert"
page_title: "ilert: ilert_status_page_group"
sidebar_current: "docs-ilert-resource-status-page-group"
description: |-
  Creates and manages a status page group in ilert.
---

# ilert_status_group

A [status page group](https://api.ilert.com/api-docs/#tag/Status-Pages) helps you organize and structure entities in your status page in collapsable containers.

> Note: Please follow our instructions for creating a status page group in a status page in README.md under examples/status_page_group within the provider.

## Example Usage

```hcl
resource "ilert_service" "example" {
  name = "example"
}

# data "ilert_status_page" "example" {
#   name = "example"
# }

# resource "ilert_status_page_group" "example" {
#   name = "example"
#   status_page {
#     id = data.ilert_status_page.example.id
#   }
# }

resource "ilert_status_page" "example" {
  name       = "example"
  subdomain  = "example.ilert.io"
  visibility = "PUBLIC"

  service {
    id = ilert_service.example.id
  }

  # structure {
  #   element {
  #     id   = ilert_status_page_group.example.id
  #     type = "GROUP"
  #     child {
  #       id   = ilert_service.example.id
  #       type = "SERVICE"
  #     }
  #   }
  #   element {
  #     id   = ilert_service.example.id
  #     type = "SERVICE"
  #   }
  # }
}

```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the status page group.
- `status_page` - (Required) A [status_page](#status-page-arguments) block. Note that only an `id` can be entered here.

#### Status Page Arguments

- `id` - (Required) The id of the status page.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the status page group.
- `name` - The name of the status page group.

## Import

Status page groups can be imported using the `id`, e.g.

```sh
$ terraform import ilert_status_page_group.main 123456789
```
