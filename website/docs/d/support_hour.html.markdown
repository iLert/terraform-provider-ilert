---
layout: "ilert"
page_title: "ilert: ilert_support_hour"
sidebar_current: "docs-ilert-data-source-support-hour"
description: |-
  Get information about a support hour that you have created.
---

# ilert_support_hour

Use this data source to get information about a specific [support hour][1].

## Example Usage

```hcl
data "ilert_support_hour" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The support hour name to use to find a support hour in the ilert API.

## Attributes Reference

- `id` - The ID of the found support hour.
- `name` - The name of the found support hour.

[1]: https://api.ilert.com/api-docs/#tag/Support-Hours
