---
layout: "ilert"
page_title: "iLert: ilert_connector"
sidebar_current: "docs-ilert-resource-connector"
description: |-
  Creates and manages a connector in iLert.
---

# ilert_connector

A [connector](https://docs.ilert.com/getting-started/intro#connectors-and-connections-outbond-integrations) is created globally in iLert and usually contains all the information to connect with the target system.

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
- `type` - (Required) The type of the connector. Allowed values are `aws_lambda`, `azure_faas`, `datadog`, `discord`, `email`, `github`, `google_faas`, `jira`, `microsoft_teams`, `servicenow`, `slack`, `sysdig`, `topdesk`, `webhook`, `zapier`, `zendesk`.
- `datadog` - (Required) A [datadog](#datadog-arguments) block.
- `jira` - (Required) A [jira](#jira-arguments) block.
- `microsoft_teams` - (Required) A [microsoft_teams](#microsoft-teams-arguments) block.
- `servicenow` - (Required) A [servicenow](#servicenow-arguments) block.
- `zendesk` - (Required) A [zendesk](#zendesk-arguments) block.
- `discord` - (Required) A [discord](#discord-arguments) block.
- `github` - (Required) A [github](#github-arguments) block.
- `topdesk` - (Required) A [topdesk](#topdesk-arguments) block.
- `aws_lambda` - (Required) A [aws_lambda](#aws-lambda-arguments) block.
- `azure_faas` - (Required) A [azure_faas](#azure-function-arguments) block.
- `google_faas` - (Required) A [google_faas](#google-cloud-function-arguments) block.
- `sysdig` - (Required) A [sysdig](#sysdig-arguments) block.

#### Datadog Arguments

> See [the Datadog outbound integration documentation](https://docs.ilert.com/integrations/datadog/outbound) for more details.

- `api_key` - (Required) The datadog API key.

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

#### AWS Lambda Arguments

> See [the AWS Lambda integration documentation](https://docs.ilert.com/integrations/aws-lambda) for more details.

- `authorization` - (Optional) The AWS Lambda authorization header value for the HTTP request.

#### Azure Function Arguments

> See [the Azure Function integration documentation](https://docs.ilert.com/integrations/azure-functions) for more details.

- `authorization` - (Optional) The Azure Function authorization header value for the HTTP request.

#### Google Cloud Function Arguments

> See [the Google Cloud Function integration documentation](https://docs.ilert.com/integrations/gcf) for more details.

- `authorization` - (Optional) The Google Function authorization header value for the HTTP request.

#### Sysdig Arguments

> See [the Sysdig outbound integration documentation](https://docs.ilert.com/integrations/sysdig/outbound) for more details.

- `api_key` - (Required) The Sysdig API key.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the connector.
- `name` - The name of the connector.
- `type` - The type of the connector.
- `created_at` - The creation date time of the connector in in ISO 8601 format.
- `updated_at` - The creation date time of the connector in in ISO 8601 format.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_connector.main 5522df22-be11-4412-ad09-5f7afbee4c2
```
