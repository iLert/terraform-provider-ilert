---
layout: "ilert"
page_title: "ilert: ilert_status_page_group"
sidebar_current: "docs-ilert-data-source-status-page-group"
description: |-
  Get information about a status page group that you have created.
---

# ilert_status_page_group

Use this data source to get information about a specific status page group on a specific [status page][1].

## Example Usage

```hcl
data "ilert_status_page_group" "example" {
  name = "example"
  status_page {
    id = 10000  # your status page id
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The status page group name to use to find a status page group on a status page in the ilert API.
- `status_page` - (Required) A [status_page](#status-page-arguments) block. Note that only an `id` can be entered here.

#### Status Page Arguments

- `id` - (Required) The id of the status page.

## Attributes Reference

- `id` - The ID of the found status_page_group.
- `name` - The name of the found status_page_group.

[1]: https://api.ilert.com/api-docs/#tag/status_page_groups
