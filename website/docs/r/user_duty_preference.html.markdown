---
layout: "ilert"
page_title: "ilert: ilert_user_duty_preference"
sidebar_current: "docs-ilert-resource-user-duty-preference"
description: |-
  Creates and manages a user duty preference in ilert.
---

# ilert_user_duty_preference

A [user duty preference](https://api.ilert.com/api-docs/#tag/Notification-Preferences) is one of the possible notification methods and defines when and how a user is getting notified when the user is about to be on-call.

## Example Usage

```hcl
resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}

data "ilert_user_email_contact" "example" {
  target = "example@example.com"
  user {
    id = ilert_user.example.id
  }
}

resource "ilert_user_duty_preference" "example" {
  method = "EMAIL"
  contact {
    id = data.ilert_user_email_contact.example.id
  }
  before_min = 0
  type       = "ON_CALL"
  user {
    id = ilert_user.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `method` - (Required) The method of the user duty preference. Allowed values are `EMAIL`, `SMS`, `VOICE`, `WHATSAPP`, `TELEGRAM`.
- `contact` - (Optional) A [contact](#contact-arguments) block. Required when `method` is `EMAIL`, `SMS`, `VOICE`, `WHATSAPP`, `TELEGRAM`. Must not be set when `method` is `PUSH`.
- `before_min` - (Required) Determines how many minutes in advance the notification should happen. Allowed values are `0`, `15`, `30`, `60`, `180`, `360`, `720`, `1440`.
- `type` - (Required) The notification type of the user duty preference. Allowed values are `ON_CALL`.
- `user` - (Required) A [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

#### Contact Arguments

- `id` - (Required) The ID of the user contact. Must be either a `UserEmailContact` or a `UserPhoneNumberContact`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user duty preference.
- `method` - (Required) The method of the user duty preference. Allowed values are `EMAIL`, `SMS`, `PUSH`, `WHATSAPP`, `TELEGRAM`.
- `contact` - (Required) A [contact](#contact-arguments) block.
- `before_min` - (Required) Determines how many minutes in advance the notification should happen. Allowed values are `0`, `15`, `30`, `60`, `180`, `360`, `720`, `1440`.
- `type` - (Required) The notification type of the user duty preference. Allowed values are `ON_CALL`.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user_duty_preference.main 123456789
```
