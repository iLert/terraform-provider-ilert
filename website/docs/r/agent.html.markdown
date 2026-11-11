---
layout: "ilert"
page_title: "ilert: ilert_agent"
sidebar_current: "docs-ilert-resource-agent"
subcategory: "Early access"
description: |-
  Creates and manages an agent in ilert.
---

# ilert_agent

An agent is an AI agent in ilert. It acts through the tools of the MCP servers it is given, which reach other systems with the [credentials](credential.html) they authenticate with.

~> **Early access:** agents are an [early access](https://docs.ilert.com/developer-docs/rest-api/api-reference/introduction/early-access) resource of the ilert API, and disabled by default: add `"ilert_agent"` to `early_access_enabled` in the [provider configuration](../index.html#early-access) to use this resource and its data source. The provider sends the `ilert-beta` header the API requires for them. They may change or be removed, and so may this resource, even in a minor version of the provider.

## Example Usage

```hcl
provider "ilert" {
  early_access_enabled = ["ilert_agent", "ilert_credential"]
}

resource "ilert_credential" "github" {
  name         = "github"
  type         = "BEARER"
  secret_value = var.github_token
  category     = "MCP"
  sub_category = "github"
}

resource "ilert_agent" "triage" {
  name = "triage"

  team {
    id = ilert_team.ops.id
  }

  mcp_server {
    name          = "ilert-workforce-github-mcp"
    credential_id = ilert_credential.github.id
  }

  mcp_tool {
    server = "ilert-workforce-github-mcp"
    name   = "get_pull_request"

    # the agent always looks at the iLert organization
    input_overwrite {
      name   = "owner"
      values = ["iLert"]
    }

    # and chooses one of these repositories
    input_overwrite {
      name   = "repo"
      values = ["ilert-go", "terraform-provider-ilert"]
    }
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the agent, unique within the account.
- `status` - (Optional) Whether the agent is active. Allowed values are `ENABLED` and `DISABLED`. Default: `ENABLED`.
- `avatar_url` - (Optional) The URL of the image the agent is shown with. Left out, the avatar set in the web app is kept. Terraform reads an empty string as left out, so removing the avatar takes the web app.
- `team` - (Optional) One or more [team](#team-arguments) blocks. The order in which the blocks are declared is not significant. Teams are managed by Terraform only once a block is declared.
- `mcp_server` - (Optional) One or more [MCP server](#mcp-server-arguments) blocks, the servers the agent can call tools on. Like teams, the servers are managed by Terraform only once a block is declared: a configuration without one leaves the servers set in the web app as they are, and removing every block after one was declared removes them all.
- `mcp_tool` - (Optional) One or more [MCP tool](#mcp-tool-arguments) blocks, the tools of the servers the agent may call. Managed only once declared, like `mcp_server`.

The names of the servers and of their tools are listed by the API at `GET /api/beta/agentic-workforce/metadata/servers-and-tools`. An agent set up in the web app and exported to Terraform from there carries them too.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### MCP Server Arguments

- `name` - (Required) The name of the server. A built-in server is referenced by its name, such as `ilert-workforce-github-mcp`, not by its short name `github`. A custom server takes a name of its own.
- `active` - (Optional) Whether the agent uses the server. Default: `true`.
- `credential_id` - (Optional) The ID of the [credential](credential.html) the server authenticates with. The credential has to be in the `MCP` category and readable with the API token of the provider. A custom server takes its URL from the credential, whose `sub_category` is `custom`.
- `display_name` - (Optional) The name a custom server is shown with.
- `brand` - (Optional) The icon a custom server is shown with, for example `aws`, `datadog` or `generic`.

#### MCP Tool Arguments

- `server` - (Required) The name of the server the tool belongs to, as in `mcp_server`.
- `name` - (Required) The name of the tool.
- `active` - (Optional) Whether the agent may call the tool. Default: `true`.
- `input_overwrite` - (Optional) One or more [input overwrite](#input-overwrite-arguments) blocks. An input without one is left to the agent.

#### Input Overwrite Arguments

- `name` - (Required) The name of the input of the tool.
- `values` - (Required) The values of the input, written as strings. The agent uses a single value as it is and chooses from several.
- `type` - (Optional) The type the values are converted to. Allowed values are `string`, `number`, `boolean`, `any` and `null`. Default: `string`.

An overwrite the web app writes without values has no effect on the agent, and is left out of the state.

### What the provider leaves alone

The API replaces an agent with what an update sends and clears most of what the update leaves out. The provider therefore reads the agent before every update and sends back what it does not manage as it was: the avatar and, while they are not declared, the teams, MCP servers and tools, the MCP servers and tools exactly as the API returned them. The settings the resource has no argument for are sent back the same way, such as the additional instructions, the autonomy level and the policy for root cause investigations, which the API only takes from accounts in the beta program.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the agent.
- `name` - The name of the agent.
- `created_at` - The creation date of the agent in ISO format.
- `updated_at` - The date the agent was last updated, in ISO format.

## Import

Agents can be imported using the `id`, e.g.

```sh
$ terraform import ilert_agent.main 123456789
```

`team`, `mcp_server` and `mcp_tool` are not captured by the import: they are managed only once declared, and the import has no configuration to read that from. The first plan after importing applies whatever the configuration declares.

Note that the API answers a read of an agent your API token cannot see with `404` rather than `403`, which is indistinguishable from a deleted one. Refreshing with a token that lacks access to the agent, one scoped to teams the agent is not assigned to for example, drops it from the state, and the next apply fails to create it again because the name is taken. Use a token with access to every agent and credential the configuration manages.
