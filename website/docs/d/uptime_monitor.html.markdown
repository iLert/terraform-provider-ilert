---
layout: "ilert"
page_title: "ilert: ilert_uptime_monitor"
sidebar_current: "docs-ilert-data-source-uptime-monitor"
description: |-
  Get information about an uptime monitor that you have created.
---

# ilert_uptime_monitor

Use this data source to get information about a specific [uptime monitor][1].

## Example Usage

```hcl
data "ilert_uptime_monitor" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The uptime monitor name to use to find a uptime monitor in the ilert API.

## Attributes Reference

- `id` - The ID of the found uptime monitor.
- `name` - The name of the found uptime monitor.
- `status` - The status of the found uptime monitor.
- `embed_url` - The embed report url of the found uptime monitor.
- `shared_url` - The shared report url of the found uptime monitor.

[1]: https://api.ilert.com/api-docs/#tag/Uptime-Monitors
