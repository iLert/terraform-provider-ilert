---
layout: "ilert"
page_title: "ilert: ilert_event_flow"
sidebar_current: "docs-ilert-data-source-event-flow"
description: |-
  Get information about an event flow that you have created.
---

# ilert_event_flow

Use this data source to get information about a specific [event flow][1].

## Example Usage

```hcl
data "ilert_event_flow" "example" {
  name = "example-event-flow"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The event flow name to use to find an event flow in the ilert API.

## Attributes Reference

- `id` - The ID of the found event flow.
- `name` - The name of the found event flow.

[1]: https://api.ilert.com/api-docs/#tag/event-flows
