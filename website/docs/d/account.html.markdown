---
layout: "ilert"
page_title: "ilert: ilert_account"
sidebar_current: "docs-ilert-data-source-account"
description: |-
  Get information about the ilert account the provider is authenticated against.
---

# ilert_account

Use this data source to get information about the [account][1] the configured credentials belong to, including the plan it is subscribed to.

## Example Usage

```hcl
data "ilert_account" "current" {}

output "account_timezone" {
  value = data.ilert_account.current.timezone
}
```

## Argument Reference

This data source takes no arguments, it always describes the account of the configured credentials.

## Attributes Reference

- `id` - The ID of the account. Unlike other entities this is a slug, not a number.
- `organization_name` - The name of the organization owning the account.
- `timezone` - The default time zone of the account as an IANA time zone name, for example `Europe/Berlin`.
- `language` - The default language of the account as an ISO 639-1 code, either `en` or `de`.
- `region` - The default country of the account as an ISO 3166-1 alpha-2 code, for example `DE`.
- `enforce_mobile_protection` - Whether users of the ilert mobile app must unlock it with an additional authentication layer.
- `allow_admin_seat_purchase` - Whether admins may purchase additional seats when creating users, rather than the account owner only.
- `allow_ai` - Whether AI features are enabled for the account.
- `ai_mode` - Where AI features of the account are allowed to run. One of `DISABLED`, `ALL` or `EU`.
- `application_features` - The features unlocked for the account by its plan and add-ons. The set of possible values grows as ilert ships new features.
- `subscription` - The plan the account is subscribed to.
  - `name` - The display name of the plan.
  - `status` - The lifecycle state of the subscription, for example `ACTIVE` or `TRIAL`.

[1]: https://docs.ilert.com/developer-docs/rest-api/api-reference/account
