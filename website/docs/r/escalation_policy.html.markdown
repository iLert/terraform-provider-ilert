---
layout: "ilert"
page_title: "ilert: ilert_escalation_policy"
sidebar_current: "docs-ilert-resource-escalation-policy"
description: |-
  Creates and manages an escalation policy in ilert.
---

# ilert_escalation_policy

An [escalation policy](https://api.ilert.com/api-docs/#tag/Escalation-Policies) connects an alert source with the users that are responsible for this alert source. It defines which users or on-call schedules should be notified when an alert is created.

## Example Usage

```hcl
resource "ilert_user" "example" {
  first_name = "example"
  last_name  = "example"
  email      = "example@example.com"
}

resource "ilert_escalation_policy" "example" {
  name = "example"

  escalation_rule {
    escalation_timeout = 15
    users {
      id = ilert_user.example.id
    }
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the escalation policy.
- `escalation_rule` - (Required) One or more [escalation rule](#escalation-rule-arguments) blocks.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `repeating` - (Optional) Indicates whether or not the escalation policy will repeat. Default: `true`.
- `frequency` - (Optional) The number of times the escalation policy will repeat after reaching the end of its escalation. This option is allowed if `repeating` is `true`. Must be between `1..9`, Default: `1`.
- `delay_min` - (Optional) Delay in minutes after which the alert gets assigned to somebody after creation. Must be between `0..15`, Default: `0`.
- `routing_key` - (Optional) The routing key of the escalation policy.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### Escalation Rule Arguments

- `escalation_timeout` - (Required) The number of minutes before an unacknowledged incident escalates away from this rule.
- `user` - (Optional) The user id of the escalation rule. Conflicts with `schedule`, `users` and `schedules`.
- `schedule` - (Optional) The schedule id of the escalation rule. Conflicts with `user`, `users` and `schedules`.
- `users` - (Optional) One or more [user](#user-arguments) blocks. Conflicts with `user` and `schedule`.
- `schedules` - (Optional) One or more [schedule](#schedule-arguments) blocks. Conflicts with `user` and `schedule`.

#### User Arguments

- `id` - (Required) The ID of the user.
- `first_name` - (Optional) The first name of the user.
- `last_name` - (Optional) The last name of the user.

#### Schedule Arguments

- `id` - (Required) The ID of the schedule.
- `name` - (Optional) The name of the schedule.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the escalation policy.
- `name` - The name of the escalation policy.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_escalation_policy.main 123456789
```
