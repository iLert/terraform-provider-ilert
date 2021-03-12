---
layout: "ilert"
page_title: "iLert: ilert_team"
sidebar_current: "docs-ilert-resource-team"
description: |-
  Creates and manages an team in iLert.
---

# ilert_team

A [team](https://api.ilert.com/api-docs/#tag/Teams) helps you to manage access to resources and simplify the user interface to show only the incidents and resources relevant to a team.

## Example Usage

```hcl
data "ilert_user" "example" {
  email = "example@example.com"
}

data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_team" "example" {
  name       = "My Team"
  visibility = "PRIVATE"

  member {
    user = data.ilert_user.example.id
    role = "STAKEHOLDER"
  }
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration"
  integration_type  = "GRAFANA"
  escalation_policy = data.ilert_escalation_policy.default.id
  teams             = [ilert_team.example.id]
}

```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the team.
- `visibility` - (Optional) The visibility of the team. Allowed values are `PUBLIC` and `PRIVATE`. Default: `PUBLIC`.
- `member` - (Optional) One or more [member](#member-arguments) blocks.

#### Member Arguments

- `user` - (Required) The user id of the team member.
- `role` - (Optional) The role of the team member. Allowed values are `ADMIN`, `USER`, `RESPONDER` and `STAKEHOLDER`. Default: `RESPONDER`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the team.
- `name` - The name of the team.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_team.main 123456789
```
