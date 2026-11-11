---
layout: "ilert"
page_title: "ilert: ilert_agent"
sidebar_current: "docs-ilert-data-source-agent"
subcategory: "Early access"
description: |-
  Get information about an agent that you have created.
---

# ilert_agent

Use this data source to get information about a specific [agent](../r/agent.html).

~> **Early access:** agents are an [early access](https://docs.ilert.com/developer-docs/rest-api/api-reference/introduction/early-access) resource of the ilert API, and disabled by default: add `"ilert_agent"` to `early_access_enabled` in the [provider configuration](../index.html#early-access) to use this data source and its resource. They may change or be removed, and so may this data source, even in a minor version of the provider.

## Example Usage

```hcl
data "ilert_agent" "triage" {
  name = "triage"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The agent name to use to find an agent in the ilert API. A name with a `/` cannot be looked up, the API does not accept one in the lookup.

## Attributes Reference

- `id` - The ID of the found agent.
- `name` - The name of the found agent.
- `status` - Whether the found agent is active, `ENABLED` or `DISABLED`.
