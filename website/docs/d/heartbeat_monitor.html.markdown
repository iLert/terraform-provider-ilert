---
layout: "ilert"
page_title: "ilert: ilert_heartbeat_monitor"
sidebar_current: "docs-ilert-data-source-heartbeat-monitor"
description: |-
    Get information about a heartbeat monitor that you have created.
---

# ilert_heartbeat_monitor

Use this data source to get information about a specific [heartbeat monitor][1].

## Example Usage

```hcl
data "ilert_heartbeat_monitor" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The heartbeat monitor name to use to find a heartbeat monitor in the ilert API.

## Attributes Reference

- `id` - The ID of the found heartbeat monitor.
- `name` - The name of the found heartbeat monitor.
- `state` - (Computed) The state of the heartbeat monitor.
- `integration_key` - (Computed) The integration key of the heartbeat monitor.
- `integration_url` - (Computed) The integration url of the heartbeat monitor.

[1]: https://api.ilert.com/api-docs/#tag/heartbeat-monitors
