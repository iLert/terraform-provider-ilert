---
layout: "ilert"
page_title: "ilert: ilert_user_email_contact"
sidebar_current: "docs-ilert-resource-user-email-contact"
description: |-
  Creates and manages a user email contact in ilert.
---

# ilert_user_email_contact

A [user email contact](https://api.ilert.com/api-docs/#tag/Contacts) is a subentity of a user and wraps various notification methods which are specifically using an email.

## Example Usage

```hcl
resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}

resource "ilert_user_email_contact" "example" {
  target = "example@example.com"
  user {
    id = ilert_user.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `target` - (Required) The target email of the user email contact.
- `user` - (Optional) The [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user email contact.
- `target` - The target email of the user email contact.
- `status` - The status of the user email contact. Possible values are: `OK`, `LOCKED`, `BLACKLISTED`.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user_email_contact.main 123456789
```
