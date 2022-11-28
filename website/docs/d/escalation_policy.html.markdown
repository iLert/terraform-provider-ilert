---
layout: "ilert"
page_title: "ilert: ilert_escalation_policy"
sidebar_current: "docs-ilert-data-source-escalation-policy"
description: |-
  Get information about an escalation policy that you have created.
---

# ilert_escalation_policy

Use this data source to get information about a specific [escalation policy][1].

## Example Usage

```hcl
data "ilert_escalation_policy" "default" {
  name = "Default"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The escalation policy name to use to find an escalation policy in the ilert API.

## Attributes Reference

- `id` - The ID of the found escalation policy.
- `name` - The name of the found escalation policy.

[1]: https://api.ilert.com/api-docs/#tag/Escalation-Policies
