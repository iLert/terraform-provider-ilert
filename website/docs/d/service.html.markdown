---
layout: "ilert"
page_title: "ilert: ilert_service"
sidebar_current: "docs-ilert-data-source-service"
description: |-
  Get information about a service that you have created.
---

# ilert_service

Use this data source to get information about a specific [service][1].

## Example Usage

```hcl
data "ilert_service" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The service name to use to find a service in the ilert API.

## Attributes Reference

- `id` - The ID of the found service.
- `name` - The name of the found service.
- `status` - The status of the found service.

[1]: https://api.ilert.com/api-docs/#tag/Services
