---
layout: "ilert"
page_title: "ilert: ilert_user"
sidebar_current: "docs-ilert-resource-user"
description: |-
  Creates and manages an user in ilert.
---

# ilert_user

An [user](https://api.ilert.com/api-docs/#tag/Users) is a member of a ilert account that has the ability to interact with alerts and other data on the account.

## Example Usage

```hcl
resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}
```

## Argument Reference

The following arguments are supported:

- `first_name` - (Required) The first name of the user.
- `last_name` - (Required) The last name of the user.
- `email` - (Required) The user's email address.
- `timezone` - (Optional) The user's timezone (IANA tz database names) e.g. `America/Los_Angeles` or `Europe/Zurich`.
- `position` - (Optional) The user's position.
- `department` - (Optional) The user's department.
- `language` - (Optional) The user's language. Allowed values are `en`, `de`.
- `role` - (Optional) The user's role. Allowed values are `ADMIN`, `USER`, `RESPONDER` or `STAKEHOLDER`. Default: `USER`
- `shift_color` - (Optional) The hex code for the user's shift color.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user.
- `email` - The user's email address of the user.
- `first_name` - The first name of the user.
- `last_name` - The last name of the user.
- `username` - The username of the user.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user.main 123456789
```
