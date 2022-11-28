---
layout: "ilert"
page_title: "ilert: ilert_connector"
sidebar_current: "docs-ilert-data-source-connector"
description: |-
  Get information about a connector that you have created.
---

# ilert_connector

Use this data source to get information about a specific [connector][1].

## Example Usage

```hcl
data "ilert_connector" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The connector name to use to find an connector in the ilert API.

## Attributes Reference

- `id` - The ID of the found connector.
- `name` - The name of the found connector.
- `type` - The type of the found connector.

[1]: https://api.ilert.com/api-docs/#tag/Connectors
