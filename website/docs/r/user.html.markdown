---
layout: "ilert"
page_title: "ilert: ilert_user"
sidebar_current: "docs-ilert-resource-user"
description: |-
  Creates and manages a user in ilert.
---

# ilert_user

A [user](https://api.ilert.com/api-docs/#tag/Users) is a member of a ilert account that has the ability to interact with alerts and other data on the account.

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
- `region` - (Optional) The user's region e.g. `EN`, `DE`.
- `role` - (Optional) The user's role. Allowed values are `ADMIN`, `USER`, `RESPONDER`, `STAKEHOLDER`, `GUEST` or `VIEWER`. Default: `USER`
- `shift_color` - (Optional) The hex code for the user's shift color.
- `send_no_invitation` - (Optional) Boolean whether an invitation email notification is sent to the user. Defaults to `false`.
- `purchase_seat` - (Optional) Boolean whether a license is bought for this user instead of checking the account's license quota. Defaults to `false`, so no configuration ever spends money unless this argument is written explicitly; creating a user beyond the licensed quota then fails with `QUOTA_EXCEEDED` and the apply stops. **Setting this to `true` buys a license and charges the account** at the price of the account's current plan, prorated for the rest of the billing period, and the license is not released when the user is deleted again. The purchase is unconditional: the API does not check whether a free seat is available first, so a user created with `purchase_seat = true` while the account still has free licenses buys one anyway. Only set it on a user that the account has no license for. The account also needs an active paid subscription and seat purchase by admins enabled in its billing settings, otherwise the apply fails. Only read when the user is created: changing it on an existing user has no effect, and removing the argument later refunds nothing.

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
