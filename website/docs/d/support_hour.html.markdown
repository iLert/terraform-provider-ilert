---
layout: "ilert"
page_title: "ilert: ilert_support_hour"
sidebar_current: "docs-ilert-data-source-support-hour"
description: |-
  Get information about a support hour that you have created.
---

# ilert_support_hour

Use this data source to get information about a specific [support hour][1].

## Example Usage

```hcl
data "ilert_support_hour" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The support hour name to use to find a support hour in the ilert API.

## Attributes Reference

- `id` - The ID of the found support hour.
- `name` - The name of the found support hour.
- `support_windows` - The complete weekly coverage of the found support hour, as a list of windows in ascending order of `from`, also when the support hour is defined with `support_days`. Each window covers the time from `from` up to `to`, not including `to`: a `to` at or before `from` wraps past Sunday midnight, and the whole week is the single window from `MONDAY 00:00` to `MONDAY 00:00`. The times are in the timezone of the support hour.
  - `from` - Where the window starts, with `day_of_week` (`MONDAY` to `SUNDAY`) and `time` (`HH:mm`).
  - `to` - Where the window ends, with `day_of_week` and `time`.

[1]: https://api.ilert.com/api-docs/#tag/Support-Hours
