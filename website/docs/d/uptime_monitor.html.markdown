---
layout: "ilert"
page_title: "ilert: ilert_uptime_monitor"
sidebar_current: "docs-ilert-data-source-uptime-monitor"
description: |-
  Discontinued: ilert no longer serves uptime monitors.
---

# ilert_uptime_monitor

This data source looked up an uptime monitor by name. ilert [discontinued uptime monitoring][1] after 30.06.2024.

> WARNING - ilert discontinued uptime monitoring after 30.06.2024 and no longer serves uptime monitors. Reading this data source fails. Remove it from your configuration, the data source will be removed in the next major version of the provider

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

[1]: https://docs.ilert.com/developer-docs/api-version-history/discontinuation-of-uptime-monitoring
