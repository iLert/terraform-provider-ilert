---
layout: "ilert"
page_title: "iLert: ilert_team"
sidebar_current: "docs-ilert-data-source-team"
description: |-
  Get information about an team that you have created.
---

# ilert_team

Use this data source to get information about a specific [team][1].

## Example Usage

```hcl
data "ilert_team" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The team name to use to find an team in the iLert API.

## Attributes Reference

- `id` - The ID of the found team.
- `name` - The name of the found team.
- `visibility` - The visibility of the found team.

[1]: https://api.ilert.com/api-docs/#tag/Teams
