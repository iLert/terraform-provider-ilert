---
layout: "ilert"
page_title: "iLert: ilert_alert_source"
sidebar_current: "docs-ilert-data-source-alert-source"
description: |-
  Get information about an alert source that you have created.
---

# ilert_alert_source

Use this data source to get information about a specific [alert source][1].

## Example Usage

```hcl
data "ilert_alert_source" "example" {
  name = "foo"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The alert source name to use to find an alert source in the iLert API.

## Attributes Reference

- `id` - The ID of the found alert source.
- `name` - The name of the found alert source.
- `status` - The status of the found alert source.
- `integration_key` - The integration key of the found alert source.

[1]: https://api.ilert.com/api-docs/#tag/Alert-Sources
