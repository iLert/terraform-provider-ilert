---
layout: "ilert"
page_title: "ilert: ilert_status_page"
sidebar_current: "docs-ilert-data-source-status-page"
description: |-
  Get information about an status page that you have created.
---

# ilert_status_page

Use this data source to get information about a specific [status page][1].

## Example Usage

```hcl
data "ilert_status_page" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The status page name to use to find an status page in the ilert API.

## Attributes Reference

- `id` - The ID of the found status page.
- `name` - The name of the found status page.

[1]: https://api.ilert.com/api-docs/#tag/Status-Pages
