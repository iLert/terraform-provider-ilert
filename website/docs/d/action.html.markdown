---
layout: "ilert"
page_title: "ilert: ilert_action"
sidebar_current: "docs-ilert-data-source-action"
subcategory: "Early access"
description: |-
  Get information about an action that you have created.
---

# ilert_action

Use this data source to get information about a specific [action](../r/action.html).

~> **Early access:** actions are an [early access](https://docs.ilert.com/developer-docs/rest-api/api-reference/introduction/early-access) resource of the ilert API, and disabled by default: add `"actions"` (for `/api/beta/actions`) to `early_access_enabled` in the [provider configuration](../index.html#early-access) to use this data source and its resource. They may change or be removed, and so may this data source, even in a minor version of the provider.

## Example Usage

```hcl
data "ilert_action" "discord" {
  name = "post high priority alerts to discord"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The action name to use to find an action in the ilert API. A name with a `/` cannot be looked up, the API does not accept one in the lookup. The alert actions the API lists among the actions are not found by name, use the [`ilert_alert_action`](alert_action.html) data source for those.

## Attributes Reference

- `id` - The ID of the found action.
- `name` - The name of the found action.
- `entity_kind` - What the found action is bound to, `ALERT` or `INCIDENT`.
- `template_key` - The template the found action is built from, empty for an action with its own steps.
- `enabled` - Whether the found action runs.
