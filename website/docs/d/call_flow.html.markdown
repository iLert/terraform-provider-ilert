---
layout: "ilert"
page_title: "ilert: ilert_call_flow"
sidebar_current: "docs-ilert-data-source-call-flow"
description: |-
  Get information about a call flow that you have created.
---

# ilert_call_flow

Use this data source to get information about a specific [call flow][1].

## Example Usage

```hcl
data "ilert_call_flow" "example" {
  name = "example-call-flow"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The call flow name to use to find a call flow in the ilert API.

## Attributes Reference

- `id` - The ID of the found call flow.
- `name` - The name of the found call flow.
- `assigned_number` - The assigned number object with `id`, `name` and nested `phone_number` containing `region_code` and `number`.

[1]: https://api.ilert.com/api-docs/#tag/call-flows
