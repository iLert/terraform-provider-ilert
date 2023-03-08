---
layout: "ilert"
page_title: "ilert: ilert_user_alert_preference"
sidebar_current: "docs-ilert-resource-user-alert-preference"
description: |-
  Creates and manages a user alert preference in ilert.
---

# ilert_user_alert_preference

A [user alert preference](https://api.ilert.com/api-docs/#tag/Notification-Preferences) is one of the possible notification methods and defines when and how a user is getting notified when an alert is created.

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

resource "ilert_user_alert_preference" "example" {
  method = "EMAIL"
  contact {
    id = ilert_user_email_contact.example.id
  }
  delay_min = 0
  type      = "HIGH_PRIORITY"
  user {
    id = ilert_user.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `method` - (Required) The method of the user alert preference. Allowed values are `EMAIL`, `SMS`, `VOICE`, `WHATSAPP`, `TELEGRAM`.
- `contact` - (Required) A [contact](#contact-arguments) block.
- `delay_min` - (Required) The delay of the notification in minutes. Must be a value between `0` and `120` (inclusive).
- `type` - (Required) The notification type of the user alert preference. Allowed values are `HIGH_PRIORITY`, `LOW_PRIORITY`.
- `user` - (Required) A [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

#### Contact Arguments

- `id` - (Required) The ID of the user contact. Must be either a `UserEmailContact` or a `UserPhoneNumberContact`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user alert preference.
- `method` - (Required) The method of the user alert preference. Allowed values are `EMAIL`, `SMS`, `VOICE`, `WHATSAPP`, `TELEGRAM`.
- `contact` - (Required) A [contact](#contact-arguments) block.
- `delay_min` - (Required) The delay of the notification in minutes. Must be a value between `0` and `120` (inclusive).
- `type` - (Required) The notification type of the user alert preference. Allowed values are `HIGH_PRIORITY`, `LOW_PRIORITY`.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user_alert_preference.main 123456789
```
