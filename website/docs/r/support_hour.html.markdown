---
layout: "ilert"
page_title: "ilert: ilert_support_hour"
sidebar_current: "docs-ilert-resource-support-hour"
description: |-
  Creates and manages a support hour in ilert.
---

# ilert_support_hour

A [support hour](https://docs.ilert.com/developer-docs/rest-api/api-reference/support-hours) lets you define the hours of the week during which support is provided, either as one window per day with `support_days`, or as any windows of the week with `support_windows`, such as split shifts, windows across midnight or the weekend, or the whole week. Used in an alert source.

## Example Usage

```hcl
resource "ilert_support_hour" "example" {
  name     = "example"
  timezone = "Europe/Berlin"
  support_days {
    monday {
      start = "08:00"
      end   = "17:00"
    }

    tuesday {
      start = "08:00"
      end   = "17:00"
    }

    wednesday {
      start = "08:00"
      end   = "17:00"
    }

    thursday {
      start = "08:00"
      end   = "17:00"
    }

    friday {
      start = "08:00"
      end   = "17:00"
    }
  }
}
```

### Support Windows Example

A split shift on Monday and on-call coverage across the weekend:

```hcl
resource "ilert_support_hour" "example_windows" {
  name     = "example_windows"
  timezone = "Europe/Berlin"

  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "09:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "12:00"
    }
  }

  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "13:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "17:00"
    }
  }

  support_windows {
    from {
      day_of_week = "FRIDAY"
      time        = "17:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "08:00"
    }
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the support hour.
- `team` - (Optional) One or more [team](#team-arguments) blocks. The order in which the blocks are declared is not significant.
- `timezone` - (Required) The timezone of the support hours (IANA tz database names) e.g. `America/Los_Angeles` or `Europe/Zurich`.
- `support_days` - (Optional) The [support days](#support-days-arguments) block of the support hours, one window per day. Exactly one of `support_days` and `support_windows` is required. An empty `support_days {}` block leaves the support hour without weekly coverage.
- `support_windows` - (Optional) Up to 84 [support window](#support-window-arguments) blocks, the complete weekly coverage of the support hours. Exactly one of `support_days` and `support_windows` is required. Use it for coverage that `support_days` cannot express: several windows a day, windows across midnight or the weekend, and the whole week.
- `exception` - (Optional) One or more [exception](#exception-arguments) blocks.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### Support Days Arguments

- `monday` - The [support day](#support-day-arguments) block of the support days.
- `tuesday` - The [support day](#support-day-arguments) block of the support days.
- `wednesday` - The [support day](#support-day-arguments) block of the support days.
- `thursday` - The [support day](#support-day-arguments) block of the support days.
- `friday` - The [support day](#support-day-arguments) block of the support days.
- `saturday` - The [support day](#support-day-arguments) block of the support days.
- `sunday` - The [support day](#support-day-arguments) block of the support days.

#### Support Day Arguments

- `start` - The start time of the support day, e.g. `08:00`. Default: `08:00`
- `end` - The end time of the support day, after `start`. The latest end is `23:59`, a support day cannot run across midnight: use `support_windows` for that. Default: `17:00`

#### Support Window Arguments

- `from` - (Required) A [time of week](#time-of-week-arguments) block, where the window starts.
- `to` - (Required) A [time of week](#time-of-week-arguments) block, where the window ends.

A window covers the time from `from` up to `to`, not including `to`. A `to` at or before `from` wraps past Sunday midnight, so `FRIDAY 17:00` to `MONDAY 08:00` is one window across the weekend, and a `to` equal to `from` covers the whole week.

Windows have to be written the way the API returns them, the plan rejects them otherwise and lists the windows to write instead:

- in ascending order of `from`, starting on Monday, so a window across Sunday midnight comes last;
- merged: windows that overlap or touch are one window, so `MONDAY 09:00` to `MONDAY 12:00` and `MONDAY 12:00` to `MONDAY 17:00` are written as `MONDAY 09:00` to `MONDAY 17:00`;
- the whole week as the single window from `MONDAY 00:00` to `MONDAY 00:00`.

#### Time of Week Arguments

- `day_of_week` - (Required) The day of the week. Allowed values are: `MONDAY`, `TUESDAY`, `WEDNESDAY`, `THURSDAY`, `FRIDAY`, `SATURDAY`, `SUNDAY`
- `time` - (Required) The time of day in `HH:mm`, e.g. `09:00`.

#### Coverage Changed Outside of Terraform

`support_days` holds one window per day. When the coverage of a support hour needs more than that, for example because a window was added in the ilert web app, the plan shows the coverage as `support_windows`, and applying the configuration replaces it with the configured one. Move such coverage into `support_windows` to keep it. Coverage that fits one window per day is kept in whichever of the two arguments the configuration uses.

#### Exception Arguments

- `name` - (Optional) The name of the exception.
- `start` - (Required) The start date and time of the exception.
- `end` - (Required) The end date and time of the exception.
- `support_status` - (Optional) The support status of the exception. Allowed values are `DURING` or `OUTSIDE`. Default: `DURING`

## Import

Support hours can be imported using the `id`, e.g.

```sh
$ terraform import ilert_support_hour.main 123456789
```

The coverage is imported as `support_days` when it fits one window per day, and as `support_windows` otherwise.
