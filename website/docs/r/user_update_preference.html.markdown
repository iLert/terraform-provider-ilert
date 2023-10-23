---
layout: "ilert"
page_title: "ilert: ilert_user_update_preference"
sidebar_current: "docs-ilert-resource-user-update-preference"
description: |-
  Creates and manages a user update preference in ilert.
---

# ilert_user_update_preference

A [user update preference](https://api.ilert.com/api-docs/#tag/Notification-Preferences) is one of the possible notification methods and defines how a user is getting notified when the status of an alert the user is assigned to changes.

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

resource "ilert_user_update_preference" "example" {
  method = "EMAIL"
  type   = "ALERT_ACCEPTED"
  contact {
    id = data.ilert_user_email_contact.example.id
  }
  user {
    id = ilert_user.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `method` - (Required) The method of the user update preference. Allowed values are `EMAIL`, `SMS`, `PUSH`.
- `type` - (Required) The notification type of the user update preference. Allowed values are `ALERT_ACCEPTED`, `ALERT_RESOLVED`, `ALERT_ESCALATED`.
- `contact` - (Optional) A [contact](#contact-arguments) block. Required when `method` is `EMAIL`, `SMS`, `VOICE`, `WHATSAPP`, `TELEGRAM`. Must not be set when `method` is `PUSH`.
- `user` - (Required) A [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

#### Contact Arguments

- `id` - (Required) The ID of the user contact. Must be either a `UserEmailContact` or a `UserPhoneNumberContact`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user update preference.
- `method` - (Required) The method of the user update preference. Allowed values are `EMAIL`, `SMS`, `VOICE`, `WHATSAPP`, `TELEGRAM`.
- `type` - (Required) The notification type of the user update preference. Allowed values are `ALERT_ACCEPTED`, `ALERT_RESOLVED`, `ALERT_ESCALATED`.
- `contact` - (Required) A [contact](#contact-arguments) block.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user_update_preference.main 123456789
```
