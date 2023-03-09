---
layout: "ilert"
page_title: "ilert: ilert_user_email_contact"
sidebar_current: "docs-ilert-data-source-user-email-contact"
description: |-
  Get information about a user email contact that you have created.
---

# ilert_user_email_contact

Use this data source to get information about a specific [user email contact][1].

## Example Usage

```hcl
data "ilert_user_email_contact" "example" {
  target = "example@example.com"
  user {
    id = 0  // id of the user where the contact should be found
  }
}
```

## Argument Reference

The following arguments are supported:

- `target` - (Required) The target email to use to find an email contact on given user.
- `user` - (Required) A [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

## Attributes Reference

- `id` - The ID of the found user email contact.
- `target` - The target email of the found user email contact.
- `status` - The status of the found user email contact. Possible values are: `OK`, `LOCKED`, `BLACKLISTED`.

[1]: https://api.ilert.com/api-docs/#tag/Contacts/paths/~1users~1{user-id}~1contacts~1emails~1{id}/get
