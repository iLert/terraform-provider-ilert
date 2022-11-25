---
layout: "ilert"
page_title: "ilert: ilert_alert_action"
sidebar_current: "docs-ilert-data-source-alert-action"
description: |-
  Get information about an alert action that you have created.
---

# ilert_alert_action

Use this data source to get information about a specific [alert_action][1].

## Example Usage

```hcl
data "ilert_alert_action" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The alert action name to use to find an alert action in the ilert API.

## Attributes Reference

- `id` - The ID of the found alert action.
- `name` - The name of the found alert action.
- `trigger_mode` - The trigger mode of the found alert action.

[1]: https://api.ilert.com/api-docs/#tag/Alert-Actions
