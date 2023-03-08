---
layout: "ilert"
page_title: "ilert: ilert_user_phone_number_contact"
sidebar_current: "docs-ilert-data-source-user-phone-number-contact"
description: |-
  Get information about a user phone number contact that you have created.
---

# ilert_user_phone_number_contact

Use this data source to get information about a specific [user phone number contact][1].

## Example Usage

```hcl
data "ilert_user_phone_number_contact" "example" {
  target = "+4915123456789"
  user {
    id = 0  // id of the user where the contact should be found
  }
}
```

> Important: Please provide a phone number with FQTN E.164 format (e.g. +49151..., not 0151...), otherwise the contact may not be found.

## Argument Reference

The following arguments are supported:

- `target` - (Required) The target phone number to use to find an phone number contact on given user.
- `user` - (Optional) A [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

## Attributes Reference

- `id` - The ID of the found user phone number contact.
- `target` - The target phone number of the found user phone number contact.
- `region_code` - The region code of the found user phone number contact.
- `status` - The status of the found user phone number contact. Possible values are: `OK`, `LOCKED`, `BLACKLISTED`.

[1]: https://api.ilert.com/api-docs/#tag/Contacts/paths/~1users~1{user-id}~1contacts~1phone-numbers~1{id}/get
