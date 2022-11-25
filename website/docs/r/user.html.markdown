---
layout: "ilert"
page_title: "ilert: ilert_user"
sidebar_current: "docs-ilert-resource-user"
description: |-
  Creates and manages an user in ilert.
---

# ilert_user

An [user](https://api.ilert.com/api-docs/#tag/Users) is a member of a ilert account that have the ability to interact with incidents and other data on the account.

## Example Usage

```hcl
resource "ilert_user" "example" {
  email      = "example@example.com"
  username   = "example"
  first_name = "example"
  last_name  = "example"

  mobile {
    region_code = "DE"
    number      = "+491234567890"
  }

  high_priority_notification_preference {
    method = "EMAIL"
    delay  = 0
  }

  low_priority_notification_preference {
    method = "EMAIL"
    delay  = 0
  }

  on_call_notification_preference {
    method     = "EMAIL"
    before_min = 60
  }
}
```

## Argument Reference

The following arguments are supported:

- `username` - (Required) The username of the user.
- `first_name` - (Required) The first name of the user.
- `last_name` - (Required) The last name of the user.
- `email` - (Required) The user's email address.
- `mobile` - (Optional) The [mobile phone number](#phone-number-arguments) block.
- `landline` - (Optional) The [landline phone number](#phone-number-arguments) block.
- `timezone` - (Optional) The user's timezone (IANA tz database names) e.g. `America/Los_Angeles` or `Europe/Zurich`.
- `position` - (Optional) The user's position.
- `department` - (Optional) The user's department.
- `language` - (Optional) The user's language. Allowed values are `en`, `de`.
- `role` - (Optional) The user's role. Allowed values are `ADMIN`, `USER`, `RESPONDER` or `STAKEHOLDER`. Default: `USER`
- `high_priority_notification_preference` - (Optional) One or more [high priority notification preference](#high-priority-notification-preference-arguments) blocks.
- `low_priority_notification_preference` - (Optional) One or more [low priority notification preference](#low-priority-notification-preference-arguments) blocks.
- `on_call_notification_preference` - (Optional) One or more [on-call notification preference](#on-call-notification-preference-arguments) blocks.
- `subscribed_incident_update_states` - (Optional) A list of subscribed incident update states. Allowed values are `ACCEPTED`, `ESCALATED` or `RESOLVED`.
- `subscribed_incident_update_notification_types` - (Optional) A list of subscribed incident update notification types. Allowed values are `EMAIL`, `ANDROID`, `IPHONE`, `SMS`, `VOICE_MOBILE`, `VOICE_LANDLINE` or `WHATSAPP`.

#### High Priority Notification Preference Arguments

- `method` - The method of the notification preference. Allowed values are `EMAIL`, `SMS`, `ANDROID`, `IPHONE`, `VOICE_MOBILE`, `VOICE_LANDLINE` or `WHATSAPP`.
- `delay` - The delay of the notification preference in minutes.

#### Low Priority Notification Preference Arguments

- `method` - The method of the notification preference. Allowed values are `EMAIL`, `SMS`, `ANDROID`, `IPHONE`, `VOICE_MOBILE`, `VOICE_LANDLINE` or `WHATSAPP`.
- `delay` - The delay of the notification preference in minutes.

#### On-Call Notification Preference Arguments

- `method` - The method of the on-call notification preference. Allowed values are `EMAIL`, `SMS`, `ANDROID`, `IPHONE` or `WHATSAPP`.
- `before_min` - The before time of the on-call notification preference in minutes.

#### Phone Number Arguments

- `region_code` - The region code e.g. `DE`
- `number` - The phone number e.g. `+4923456789`

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the user.
- `username` - The username of the user.
- `email` - The user's email address of the user.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_user.main 123456789
```
