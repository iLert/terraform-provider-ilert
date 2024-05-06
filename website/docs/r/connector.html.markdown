---
layout: "ilert"
page_title: "ilert: ilert_connector"
sidebar_current: "docs-ilert-resource-connector"
description: |-
  Creates and manages a connector in ilert.
---

# ilert_connector

A [connector](https://docs.ilert.com/getting-started/readme#connectors-and-alert-actions-aka-outbound-integrations) is created globally in ilert and usually contains all the information to connect with the target system.

## Example Usage

```hcl
resource "ilert_connector" "example" {
  name = "My GitHub Connector"
  type = "github"

  github {
    api_key = "my api key"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the connector.
- `type` - (Required) The type of the connector. Allowed values are `jira`, `microsoft_teams`, `servicenow`, `zendesk`, `discord`, `github`, `topdesk`, `autotask`, `mattermost`, `zammad`, `dingtalk`.
- `jira` - (Optional) A [jira](#jira-arguments) block.
- `microsoft_teams` - (Optional) A [microsoft_teams](#microsoft-teams-arguments) block.
- `servicenow` - (Optional) A [servicenow](#servicenow-arguments) block.
- `zendesk` - (Optional) A [zendesk](#zendesk-arguments) block.
- `discord` - (Optional) A [discord](#discord-arguments) block.
- `github` - (Optional) A [github](#github-arguments) block.
- `topdesk` - (Optional) A [topdesk](#topdesk-arguments) block.
- `autotask` - (Optional) A [autotask](#autotask-arguments) block.
- `mattermost` - (Optional) A [mattermost](#mattermost-arguments) block.
- `zammad` - (Optional) A [zammad](#zammad-arguments) block.
- `dingtalk` - (Optional) A [dingtalk](#dingtalk-arguments) block.

#### Jira Arguments

> See [the Jira outbound integration documentation](https://docs.ilert.com/integrations/jira/outbound) for more details.

- `url` - (Required) The Jira server URL.
- `email` - (Required) The Jira user email.
- `password` - (Required) The Jira user password or API token.

#### Microsoft Teams Arguments

> See [the Microsoft Teams integration documentation](https://docs.ilert.com/integrations/microsoft-teams) for more details.

- `url` - (Required) The Microsoft Teams connector URL.

#### ServiceNow Arguments

> See [the ServiceNow integration documentation](https://docs.ilert.com/integrations/service-now) for more details.

- `url` - (Required) The ServiceNow server URL.
- `username` - (Required) The ServiceNow username.
- `password` - (Required) The ServiceNow user password.

#### Zendesk Arguments

> See [the Zendesk integration documentation](https://docs.ilert.com/integrations/zendesk) for more details.

- `url` - (Required) The Zendesk server URL.
- `email` - (Required) The Zendesk user email.
- `api_key` - (Required) The Zendesk user API key.

#### Discord Arguments

> See [the Discord integration documentation](https://docs.ilert.com/integrations/discord) for more details.

- `url` - (Required) The Discord connector URL.

#### GitHub Arguments

> See [the GitHub outbound issue integration documentation](https://docs.ilert.com/integrations/github/outbound-issue) for more details.

- `api_key` - (Required) The GitHub API key.

#### TOPdesk Arguments

> See [the TOPdesk integration documentation](https://docs.ilert.com/integrations/topdesk/outbound) for more details.

- `url` - (Required) The TOPdesk server URL.
- `username` - (Required) The TOPdesk username.
- `password` - (Required) The TOPdesk user password.

#### Autotask Arguments

> See [the Autotask outbound integration documentation](https://docs.ilert.com/integrations/autotask/outbound) for more details.

- `url` - (Required) The Autotask server URL.
- `email` - (Required) The Autotask email.
- `password` - (Required) The Autotask user password.

#### Mattermost Arguments

> See [the Mattermost outbound integration documentation](https://docs.ilert.com/integrations/mattermost) for more details.

- `url` - (Required) The Mattermost server URL.

#### Zammad Arguments

> See [the Zammad outbound integration documentation](https://docs.ilert.com/integrations/zammad/outbound) for more details.

- `url` - (Required) The Zammad server URL.
- `api_key` - (Required) The Zammad API key.

#### Dingtalk Arguments

> See [the Dingtalk outbound integration documentation](https://docs.ilert.com/integrations/dingtalk) for more details.

- `url` - (Required) The Dingtalk server URL.
- `secret` - (Required) The Dingtalk secret.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the connector.
- `name` - The name of the connector.
- `type` - The type of the connector.
- `created_at` - The creation date time of the connector in ISO 8601 format.
- `updated_at` - The creation date time of the connector in ISO 8601 format.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_connector.main 123456789
```
