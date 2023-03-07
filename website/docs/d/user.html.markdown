---
layout: "ilert"
page_title: "ilert: ilert_user"
sidebar_current: "docs-ilert-data-source-user"
description: |-
  Get information about an user that you have created.
---

# ilert_user

Use this data source to get information about a specific [user][1].

## Example Usage

```hcl
data "ilert_user" "example" {
  email = "example@example.com"
}
```

## Argument Reference

The following arguments are supported:

- `email` - (Required) The user email to use to find an user in the ilert API.

## Attributes Reference

- `id` - The ID of the found user.
- `email` - The email of the found user.
- `first_name` - The first name of the found user.
- `last_name` - The last name of the found user.
- `username` - The username of the found user.

[1]: https://api.ilert.com/api-docs/#tag/Users
