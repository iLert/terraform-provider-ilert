---
layout: "ilert"
page_title: "ilert: ilert_call_flow_number"
sidebar_current: "docs-ilert-data-source-call-flow-number"
description: |-
  Get information about a call flow number of your account.
---

# ilert_call_flow_number

Use this data source to get information about a specific [call flow number][1] of your account.

These are the numbers assigned to your account, which is a different entity from the numbers ilert offers for purchase.

## Example Usage

```hcl
data "ilert_call_flow_number" "example" {
  name = "support hotline"
}

output "support_hotline" {
  value = data.ilert_call_flow_number.example.phone_number[0].number
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The call flow number name to use to find a call flow number in the ilert API. The API matches the value against the phone number first and falls back to the name, so the number itself also resolves.

## Attributes Reference

- `id` - The ID of the found call flow number.
- `name` - The name of the found call flow number.
- `state` - Whether the number is assigned to a call flow, either `AVAILABLE` or `USED`. Empty when the number was matched by its phone number rather than by its name, since the API omits the state in that case.
- `phone_number` - The phone number itself.
  - `region_code` - The region code of the phone number, for example `DE`.
  - `number` - The phone number in international format.

[1]: https://docs.ilert.com/developer-docs/rest-api/api-reference/call-flow-numbers
