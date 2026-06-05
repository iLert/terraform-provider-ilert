---
layout: "ilert"
page_title: "ilert: ilert_event_flow_integration"
sidebar_current: "docs-ilert-data-source-event-flow-integration"
description: |-
  Get information about an event flow integration that you have created.
---

# ilert_event_flow_integration

Use this data source to look up an existing [event flow integration][1] by the event flow it is attached to and its integration type.

## Example Usage

```hcl
data "ilert_event_flow" "example" {
  name = "example-event-flow"
}

data "ilert_event_flow_integration" "example_api" {
  event_flow_id    = data.ilert_event_flow.example.id
  integration_type = "API"
}
```

## Argument Reference

The following arguments are supported:

- `event_flow_id` - (Required) The ID of the event flow whose integrations to search.
- `integration_type` - (Required) The integration type to match (e.g. `API`, `EMAIL`, `GRAFANA`, `DATADOG`, ...). The first integration of this type attached to the given event flow is returned.

## Attributes Reference

- `id` - The ID of the found event flow integration.
- `integration_key` - The integration key (for email-based integrations: the inbound email address).
- `integration_url` - The fully qualified URL clients should send events to (when applicable for the integration type).

[1]: https://api.ilert.com/api-docs/#tag/event-flows
