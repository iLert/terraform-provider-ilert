---
layout: "ilert"
page_title: "ilert: ilert_telemetry_source"
sidebar_current: "docs-ilert-data-source-telemetry-source"
description: |-
  Get information about a telemetry source that you have created.
---

# ilert_telemetry_source

Use this data source to get information about a specific [telemetry source][1].

## Example Usage

```hcl
data "ilert_telemetry_source" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The telemetry source name to use to find a telemetry source in the ilert API.

## Attributes Reference

- `id` - The ID of the found telemetry source.
- `name` - The name of the found telemetry source.
- `type` - The telemetry protocol of the found telemetry source.
- `description` - The description of the found telemetry source.
- `service_name_prefix` - The prefix prepended to the services ilert creates from this source.
- `service_name_suffix` - The suffix appended to the services ilert creates from this source.
- `labels` - The labels assigned to the found telemetry source. The reserved `ilert.com/discovered-by` label the API stamps on every source is omitted.
- `status` - Whether ilert is currently receiving telemetry from the found source.
- `integration_key` - The credential the collector authenticates with. Marked sensitive, and omitted by the API for users without update permission on the telemetry source.

[1]: https://docs.ilert.com/developer-docs/rest-api/api-reference/telemetry-sources
