---
layout: "ilert"
page_title: "ilert: ilert_schedule"
sidebar_current: "docs-ilert-resource-schedule"
description: |-
  Creates and manages an on-call-schedule in ilert.
---

# ilert_schedule

A [schedule](https://api.ilert.com/api-docs/#tag/Schedules) is used to dynamically determine to whom an alert will be assigned to based on the time of the day. ilert offers two types of schedules - recurring and static schedules, which differ in the way a schedule is created and maintained.

## Example Usage

```hcl
resource "ilert_user" "example" {
  first_name = "example"
  last_name  = "example"
  email      = "example@example.com"
}

# example for recurring schedule
resource "ilert_schedule" "example_recurring" {
  name     = "example_recurring"
  timezone = "Europe/Berlin"
  type     = "RECURRING"
  schedule_layer {
    name      = "layer1"
    starts_on = "2023-08-30T00:00"
    user {
      id = ilert_user.example.id
    }
    rotation         = "P1D"
    restriction_type = "TIMES_OF_WEEK"
    restriction {
      from {
        day_of_week = "MONDAY"
        time        = "13:00"
      }
      to {
        day_of_week = "MONDAY"
        time        = "16:00"
      }
    }
  }
}

# example for static schedule
resource "ilert_schedule" "example_static" {
  name     = "example_static"
  timezone = "Europe/Berlin"
  type     = "STATIC"
  shift {
    user  = ilert_user.example.id
    start = "2023-09-01T08:00"
    end   = "2023-09-02T08:00"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the schedule.
- `timezone` - (Required) The current timezone, for ex. `Europe/Berlin`.
- `type` - (Required) The type of the schedule. Allowed values are `STATIC` or `RECURRING`.
- `schedule_layer` - (Optional, type = `RECURRING`) - One or more [schedule layer](#schedule-layer-arguments) blocks.
- `shift` - (Optional, type = `STATIC`) - One or more [shift](#shift-arguments) blocks.
- `show_gaps` - (Optional) Indicates whether gaps between shifts should be shown. Default: `true`
- `default_shift_duration` - (Optional) The default furation of a shift.
- `current_shift` - (Optional) A [shift](#shift-arguments) block.
- `next_shift` - (Optional) A [shift](#shift-arguments) block.
- `team` - (Optional) One or more [team](#team-arguments) blocks.

#### Schedule Layer Arguments

- `name` - (Required) The name of the schedule layer.
- `starts_on` - (Required) The starting date and time of the schedule layer as a date time string in ISO format. For ex. `2022-08-30T00:00`
- `ends_on` - (Optional) The starting date and time of the schedule layer as a date time string in ISO format. For ex. `2022-08-30T00:00`
- `user` - (Required) One or more [user](#user-arguments) blocks.
- `rotation` - (Optional) The duration of the schedule per user in ISO format. For ex. `P7D` (7 Days) or `PT8H` (8 Hours)
- `restriction_type` - (Optional) The type of time restrictions. Allowed values are: `TIMES_OF_WEEK`
- `restriction` - (Optiomal) One or more [restriction](#restriction-arguments) blocks.

#### User Arguments

- `id` - (Required) The ID of the user.
- `first_name` - (Optional) The first name of the user.
- `last_name` - (Optional) The last name of the user.

#### Restriction Arguments

- `from` - (Required) A [time of week](#time-of-week-arguments) block.
- `to` - (Required) A [time of week](#time-of-week-arguments) block.

#### Time of week Arguments

- `day_of_week` - (Required) The day of the week. Allowed values are: `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, `SATURDAY`, `SUNDAY`
- `time`- (Required) The time on each day in a time string in format. For ex. `15:00`

#### Shift Arguments

- `user` - (Required) The ID of the user.
- `start` - (Required) The start of the shift as a date time string in ISO format. For ex. `2022-08-30T00:00`
- `end` - (Required) The start of the shift as a date time string in ISO format. For ex. `2022-08-30T00:00`

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the schedule.
- `name` - The name of the schedule.

## Import

Schedules can be imported using the `id`, e.g.

```sh
$ terraform import ilert_schedule.main 123456789
```
