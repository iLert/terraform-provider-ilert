---
layout: "ilert"
page_title: "ilert: ilert_user_phone_number_contact"
sidebar_current: "docs-ilert-resource-user-phone-number-contact"
description: |-
  Creates and manages a user phone number contact in ilert.
---

# ilert_user_phone_number_contact

A [user phone number contact](https://api.ilert.com/api-docs/#tag/Contacts) is a subentity of a user and wraps various notification methods which are specifically using a phone number.

## Example Usage

```hcl
resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}

resource "ilert_user_phone_number_contact" "example" {
  region_code = "DE"
  target      = "+4915123456789" // for best practice, use FQTN E.164 format
  user {
    id = ilert_user.example.id
  }
}
```

> Info: For best practice use a phone number in FQTN E.164 format (e.g. +49151..., not 0151...)

## Argument Reference

The following arguments are supported:

- `target` - (Required) The target phone number of the user phone number contact.
- `region_code` - (Required) The region code for the target phone number of a user phone number contact.
- `user` - (Optional) A [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user phone number contact.
- `target` - The target phone number of the user phone number contact.
- `region_code` - The region code for the target phone number of a user phone number contact.
- `status` - The status of the user phone number contact. Possible values are: `OK`, `LOCKED`, `BLACKLISTED`.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user_phone_number_contact.main 123456789
```
