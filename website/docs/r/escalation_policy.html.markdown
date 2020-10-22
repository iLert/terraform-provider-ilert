---
layout: "ilert"
page_title: "iLert: ilert_escalation_policy"
sidebar_current: "docs-ilert-resource-escalation policy"
description: |-
  Creates and manages an escalation policy in iLert.
---

# ilert_escalation_policy

An [escalation policy](https://api.ilert.com/api-docs/#tag/Escalation-Policies) determines what user or schedule will be notified first, second, and so on when an incident is triggered. Escalation policies are used by one or more alert sources.

## Example Usage

```hcl
data "ilert_user" "example" {
  email = "example@example.com"
}

data "ilert_schedule" "example" {
  name = "example"
}

resource "ilert_escalation_policy" "example" {
  name = "example"

  escalation_rule {
    escalation_timeout = 5
    schedule           = data.ilert_schedule.example.id
  }

  escalation_rule {
    escalation_timeout = 15
    user               = data.ilert_user.example.id
  }
}


```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the escalation policy.
- `repeating` - (Optional) Indicates whether or not the escalation policy will repeat. Default: `true`.
- `frequency` - (Optional) The number of times the escalation policy will repeat after reaching the end of its escalation. This option is allowed if `repeating` is `true`. Default: `1`.
- `escalation_rule` - (Optional) One or more [escalation rule](#escalation-rule-arguments) blocks.

#### Escalation Rule Arguments

- `escalation_timeout` - (Required) The number of minutes before an unacknowledged incident escalates away from this rule.
- `user` - (Optional) The user id of the escalation rule. Conflicts with `schedule`.
- `schedule` - (Optional) The schedule id of the escalation rule. Conflicts with `user`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the escalation policy.
- `name` - The name of the escalation policy.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_escalation_policy.main 123456789
```
