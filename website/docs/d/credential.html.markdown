---
layout: "ilert"
page_title: "ilert: ilert_credential"
sidebar_current: "docs-ilert-data-source-credential"
subcategory: "Early access"
description: |-
  Get information about a credential that you have created.
---

# ilert_credential

Use this data source to get information about a specific [credential](../r/credential.html), for example to attach a credential created in the web app to an agent.

~> **Early access:** credentials are an [early access](https://docs.ilert.com/developer-docs/rest-api/api-reference/introduction/early-access) resource of the ilert API, and disabled by default: add `"ilert_credential"` to `early_access_enabled` in the [provider configuration](../index.html#early-access) to use this data source and its resource. They may change or be removed, and so may this data source, even in a minor version of the provider.

## Example Usage

```hcl
data "ilert_credential" "github" {
  name = "github"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The credential name to use to find a credential in the ilert API. A name with a `/` cannot be looked up, the API does not accept one in the lookup.

## Attributes Reference

- `id` - The ID of the found credential.
- `name` - The name of the found credential.
- `type` - How the secret of the found credential is presented, for example `BEARER`.
- `host_url` - The target the found credential is used against. For a credential of type `NONE` only the scheme, host and port, its path and query are the secret.
- `category` - The category of the found credential, `MCP` for the credentials of MCP servers.
- `sub_category` - The sub category of the found credential.

The secret is never returned.
