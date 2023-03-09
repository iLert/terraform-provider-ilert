---
layout: "ilert"
page_title: "ilert: ilert_user_subscription_preference"
sidebar_current: "docs-ilert-resource-user-subscription-preference"
description: |-
  Creates and manages a user subscription preference in ilert.
---

# ilert_user_subscription_preference

A [user subscription preference](https://api.ilert.com/api-docs/#tag/Notification-Preferences) is one of the possible notification methods and defines how a user is getting notified when the user is added as a subscriber to an incident, a service or a status page.

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

resource "ilert_user_subscription_preference" "example" {
  method = "EMAIL"
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

- `method` - (Required) The method of the user subscription preference. Allowed values are `EMAIL`, `SMS`, `PUSH`.
- `contact` - (Required) A [contact](#contact-arguments) block.
- `user` - (Required) A [user](#user-arguments) block.

#### User Arguments

- `id` - (Required) The ID of the user.

#### Contact Arguments

- `id` - (Required) The ID of the user contact. Must be either a `UserEmailContact` or a `UserPhoneNumberContact`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user subscription preference.
- `method` - (Required) The method of the user subscription preference. Allowed values are `EMAIL`, `SMS`, `VOICE`, `WHATSAPP`, `TELEGRAM`.
- `contact` - (Required) A [contact](#contact-arguments) block.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user_subscription_preference.main 123456789
```
