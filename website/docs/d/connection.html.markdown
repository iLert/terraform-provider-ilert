---
layout: "ilert"
page_title: "ilert: ilert_connection"
sidebar_current: "docs-ilert-data-source-connection"
description: |-
  Get information about a connection that you have created.
---

# ilert_connection

Use this data source to get information about a specific [connection][1].

## Example Usage

```hcl
data "ilert_connection" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The connection name to use to find an connection in the ilert API.

## Attributes Reference

- `id` - The ID of the found connection.
- `name` - The name of the found connection.
- `trigger_mode` - The trigger mode of the found connection.

[1]: https://api.ilert.com/api-docs/#tag/Connectors
