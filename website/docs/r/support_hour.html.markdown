---
layout: "ilert"
page_title: "ilert: ilert_support_hour"
sidebar_current: "docs-ilert-resource-support-hour"
description: |-
  Creates and manages a support hour in ilert.
---

# ilert_support_hour

A [support hour](https://api.ilert.com/api-docs/#tag/Support-Hours) lets you define the support hours for each day of the week. Used in an alert source.

## Example Usage

```hcl
resource "ilert_support_hour" "example" {
  name = "example"
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

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the support hour.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `timezone` - (Optional) The timezone of the support hours (IANA tz database names) e.g. `America/Los_Angeles` or `Europe/Zurich`.
- `support_days` - The [support days](#support-days-arguments) block of the support hours.

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

- `start` - The start time of the support day.
- `end` - The end time of the support day.

## Import

Support hours can be imported using the `id`, e.g.

```sh
$ terraform import ilert_support_hour.main 123456789
```
