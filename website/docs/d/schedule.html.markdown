---
layout: "ilert"
page_title: "iLert: ilert_schedule"
sidebar_current: "docs-ilert-data-source-schedule"
description: |-
  Get information about an schedule that you have created.
---

# ilert_schedule

Use this data source to get information about a specific [schedule][1].

## Example Usage

```hcl
data "ilert_schedule" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The schedule name to use to find a schedule in the iLert API.

## Attributes Reference

- `id` - The ID of the found schedule.
- `name` - The name of the found schedule.

[1]: https://api.ilert.com/api-docs/#tag/Schedules
