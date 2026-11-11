---
layout: "ilert"
page_title: "ilert: ilert_credential"
sidebar_current: "docs-ilert-resource-credential"
subcategory: "Early access"
description: |-
  Creates and manages a credential in ilert.
---

# ilert_credential

A credential holds a secret that ilert uses against another system, such as the token the MCP server of an [agent](agent.html) authenticates with. ilert stores the secret encrypted and never returns it.

~> **Early access:** credentials are an [early access](https://docs.ilert.com/developer-docs/rest-api/api-reference/introduction/early-access) resource of the ilert API, and disabled by default: add `"ilert_credential"` to `early_access_enabled` in the [provider configuration](../index.html#early-access) to use this resource and its data source. The provider sends the `ilert-beta` header the API requires for them. They may change or be removed, and so may this resource, even in a minor version of the provider.

~> **A credential cannot be changed.** The API rejects every update, so changing any argument replaces the credential: Terraform deletes it and creates a new one with a new `id`. Names are unique, so the new credential can only be created once the old one is deleted, and whatever uses the credential is without it in between. Agents that reference it as `ilert_credential.<name>.id` are switched to the new one in the same apply. The API does not check whether a credential is still in use when it is deleted.

## Example Usage

```hcl
provider "ilert" {
  early_access_enabled = ["ilert_credential"]
}

# the credential of a built-in MCP server
resource "ilert_credential" "github" {
  name         = "github"
  type         = "BEARER"
  secret_value = var.github_token
  category     = "MCP"
  sub_category = "github"
}

# the credential of a custom MCP server, which takes its URL from it
resource "ilert_credential" "internal_tools" {
  name         = "internal tools"
  type         = "BEARER"
  secret_value = var.internal_tools_token
  host_url     = "https://mcp.example.com/mcp"
  category     = "MCP"
  sub_category = "custom"
}
```

## Argument Reference

The following arguments are supported. Changing any of them replaces the credential.

- `name` - (Required) The name of the credential, unique within the account.
- `type` - (Required) How the secret is presented to the system it is used against. Allowed values are `NONE`, `BEARER`, `API_KEY`, `BASIC`, `OAUTH2`, `AWS_SIGV4` and `UNKNOWN`. The other arguments a type takes are listed under [Types](#types).
- `secret_value` - (Optional) The secret, required by every type but `NONE` to create the credential. It is only sent when the credential is created and the API never returns it, so the state keeps the configured value and a secret changed outside of Terraform does not show in the plan. Marked sensitive. An imported credential can do without one, see [Import](#import).
- `username` - (Optional) The part of the secret that is not secret, see [Types](#types).
- `host_url` - (Optional) The target the secret is used against, an absolute http or https URL. A credential in the `MCP` category needs a public https host. A credential of type `NONE` takes its whole URL here, which is why the argument is marked sensitive.
- `idp_url` - (Optional) The token endpoint of an `OAUTH2` credential.
- `idp_client` - (Optional) The client id at the token endpoint.
- `idp_scopes` - (Optional) The space separated scopes requested at the token endpoint.
- `audience` - (Optional) The audience requested at the token endpoint.
- `oauth_method` - (Optional) The client authentication at the token endpoint. `CLIENT_SECRET_BASIC` and `CLIENT_SECRET_POST` use the client credentials grant, `REFRESH_TOKEN` uses the refresh token grant and sends no client secret.
- `category` - (Optional) What the credential is used for. Agents only take credentials in the `MCP` category for their MCP servers.
- `sub_category` - (Optional) For an `MCP` credential, the short name of the built-in server it belongs to, such as `github`, or `custom` for a custom server, which takes its URL from `host_url`.
- `api_key_header_name` - (Optional) The header an `API_KEY` credential is sent in.
- `role_arn` - (Optional) The role an `AWS_SIGV4` credential assumes.
- `region` - (Optional) The region of an `AWS_SIGV4` credential.
- `team` - (Optional) One or more [team](#team-arguments) blocks. The order in which the blocks are declared is not significant. Teams are managed by Terraform only once a block is declared. A team `name` is kept as configured, so renaming the team in ilert does not replace the credential.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

### Types

| `type`      | `secret_value`                                                           | `username`                                 | Also takes                                    |
| ----------- | ------------------------------------------------------------------------ | ------------------------------------------ | --------------------------------------------- |
| `NONE`      | none, the secret is the path and query of `host_url`                     | none                                       | `host_url` with the whole URL, required        |
| `BEARER`    | the token                                                                | none                                       |                                               |
| `API_KEY`   | the key, sent as it is                                                   | optional                                   | `api_key_header_name`                         |
| `BASIC`     | the password                                                             | required                                   |                                               |
| `OAUTH2`    | the client secret, or the refresh token with `oauth_method = "REFRESH_TOKEN"` | none                                  | `idp_url`, required, and `idp_client` with the client credentials grant |
| `AWS_SIGV4` | the secret access key                                                    | the access key id, unless `role_arn` is set | `role_arn`, `region`                          |
| `UNKNOWN`   | the secret, stored as it is                                              | none                                       |                                               |

A credential of type `NONE` is a URL that carries its own secret, such as a webhook URL with a token in its path. ilert keeps the path and query as the secret and returns only the scheme, host and port. The state keeps the configured URL as long as the returned origin matches it.

These rules, and the URLs the API cannot parse, are checked at plan time whenever a credential is created or replaced. The API rejects a credential that breaks them, and since a change replaces the credential, it would only fail after Terraform deleted the one it replaces. The API also requires the `host_url` and `idp_url` of an `MCP` credential to be a public https host. That rule depends on the account and is only checked by the API.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the credential.
- `name` - The name of the credential.

## Import

Credentials can be imported using the `id`, e.g.

```sh
$ terraform import ilert_credential.main 123456789
```

The secret is not imported, since the API never returns it, and neither is the whole URL of a `NONE` credential, of which it only returns the origin. To take over a credential as it is, leave `secret_value` out of the configuration, and for `NONE` set `host_url` to the origin only: the plan after the import is then empty. Setting either afterwards replaces the credential with one created from the configuration, since the secret of an existing credential cannot be changed.

Unlike on other resources, `team` is captured by the import, with the team names, as the configuration the web app exports to Terraform declares them. A team declared without its `name` replaces the credential on the first apply.

Note that the API answers a read of a credential your API token cannot see with `404` rather than `403`, which is indistinguishable from a deleted one. Refreshing with a token that lacks access to the credential, one scoped to teams the credential is not assigned to for example, drops it from the state, and the next apply fails to create it again because the name is taken. Use a token with access to every credential the configuration manages.
