---
layout: "ilert"
page_title: "ilert: ilert_alert_action"
sidebar_current: "docs-ilert-resource-alert-action"
description: |-
  Creates and manages a alert_action in ilert.
---

# ilert_alert_action

An [alert_action](https://docs.ilert.com/getting-started/readme#connectors-and-alert-actions-aka-outbound-integrations) is created at the alert source level and uses its [connector](connector.html) to perform a concrete action.

## Example Usage

```hcl
resource "ilert_user" "example" {
  username   = "example1"
  first_name = "example"
  last_name  = "example"
  email      = "example1@example.com"
}

resource "ilert_escalation_policy" "example" {
  name = "example"
  escalation_rule {
    escalation_timeout = 15
    user               = ilert_user.example.id
  }
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration for GitHub"
  integration_type  = "GRAFANA"
  escalation_policy = ilert_escalation_policy.default.id
}

resource "ilert_connector" "example" {
  name = "My GitHub Connector"
  type = "github"

  github {
    api_key = "my api key"
  }
}

resource "ilert_alert_action" "example" {
  name = "My GitHub Alert Action"

  alert_source {
    id = ilert_alert_source.example.id
  }

  connector {
    id   = ilert_connector.example.id
    type = ilert_connector.example.type
  }

  github {
    owner      = "my org"
    repository = "my repo"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the alert action.
- `alert_source` - (Required) An [alert source](#alert-source-arguments) block.
- `connector` - (Required) A [connector](#connector-arguments) block.
- `trigger_mode` - (Optional) The trigger mode of the alert action. Allowed values are `AUTOMATIC` or `MANUAL`. Default: `AUTOMATIC`.
- `trigger_types` - (Optional if the `MANUAL` trigger mode and required if the `AUTOMATIC` trigger mode ) A list of the trigger types. Allowed values are `alert-created`, `alert-assigned`, `alert-auto-escalated`, `alert-acknowledged`, `alert-raised`, `alert-comment-added`, `alert-resolved`.
- `alert_filter` - (Optional) An [alert filter](#alert-filter-arguments) block.
- `datadog` - (Optional) A [datadog](#datadog-arguments) block.
- `jira` - (Optional) A [jira](#jira-arguments) block.
- `servicenow` - (Optional) A [servicenow](#servicenow-arguments) block.
- `slack` - (Optional) A [slack](#slack-arguments) block.
- `webhook` - (Optional) A [webhook](#webhook-arguments) block.
- `zendesk` - (Optional) A [zendesk](#zendesk-arguments) block.
- `github` - (Optional) A [github](#github-arguments) block.
- `aws_lambda` - (Optional) A [aws_lambda](#aws-lambda-arguments) block.
- `azure_faas` - (Optional) A [azure_faas](#azure-function-arguments) block.
- `google_faas` - (Optional) A [google_faas](#google-cloud-function-arguments) block.
- `email` - (Optional) A [email](#email-arguments) block.
- `sysdig` - (Optional) A [sysdig](#sysdig-arguments) block.
- `zapier` - (Optional) A [zapier](#zapier-arguments) block.
- `autotask` - (Optional) A [autotask](#autotask-arguments) block.
- `mattermost` - (Optional) A [mattermost](#mattermost-arguments) block.
- `zammad` - (Optional) A [zammad](#zammad-arguments) block.
- `status_page_io` - (Optional) A [status_page_io](#statuspage-arguments) block.

#### Alert Source Arguments

- `id` - (Required) The alert source id.

#### Connector Arguments

- `id` - (Optional) The connector id. Required if the connector `type` is one of values `aws_lambda`, `azure_faas`, `datadog`, `discord`, `github`, `google_faas`, `jira`, `microsoft_teams`, `servicenow`, `sysdig`, `topdesk`, `zendesk`, `autotask`, `mattermost`, `zammad`, `status_page_io`.
- `type` - (Required) The connector type. Allowed values are `aws_lambda`, `azure_faas`, `datadog`, `discord`, `email`, `github`, `google_faas`, `jira`, `microsoft_teams`, `servicenow`, `slack`, `sysdig`, `topdesk`, `webhook`, `zapier`, `zendesk`.

#### Alert Filter Arguments

- `operator` - (Required) The operator to use for the filter. Allowed values are `AND` or `OR`.
- `predicate` - (Required) One or more [predicate](#predicate-arguments) blocks.

#### Predicate Arguments

- `field` - (Required) The field which should be monitored for conditional execution. Allowed values are `ALERT_SUMMARY`, `ALERT_DETAILS`, `ESCALATION_POLICY`, `ALERT_PRIORITY`.
- `criteria` - (Required) The criteria for the condition. Allowed values are `CONTAINS_ANY_WORDS`, `CONTAINS_NOT_WORDS`, `CONTAINS_STRING`, `CONTAINS_NOT_STRING`, `IS_STRING`, `IS_NOT_STRING`, `MATCHES_REGEX`, `MATCHES_NOT_REGEX`.
- `value` - (Required) The value for the condition.

#### Datadog Arguments

> See [the Datadog outbound integration documentation](https://docs.ilert.com/integrations/datadog/outbound) for more details.

- `priority` - (Optional) The datadog priority.
- `site` - (Optional) The datadog site. Allowed values are `EU` or `US`. Default: `EU`.
- `tags` - (Optional) A list of the datadog tags.

#### Jira Arguments

> See [the Jira outbound integration documentation](https://docs.ilert.com/integrations/jira/outbound) for more details.

- `project` - (Required) The Jira project.
- `issue_type` - (Optional) The Jira issue type. Allowed values are `Bug`, `Epic`, `Subtask`, `Story`, `Task`.
  Default: `Task`.
- `body_template` - (Optional) The Jira issue template body.

#### ServiceNow Arguments

> See [the ServiceNow integration documentation](https://docs.ilert.com/integrations/service-now) for more details.

- `caller_id` - (Optional) The ServiceNow caller id.
- `impact` - (Optional) The ServiceNow impact.
- `urgency` - (Optional) The ServiceNow urgency.

#### Slack Arguments

> See [the Slack integration documentation](https://docs.ilert.com/integrations/slack) for more details.

- `channel_id` - (Required) The Slack channel id.
- `channel_name` - (Optional) The Slack channel name.
- `team_id` - (Required) The Slack workspace id.
- `team_domain` - (Optional) The Slack workspace name.

#### Webhook Arguments

> See [the Webhook integration documentation](https://docs.ilert.com/integrations/webhook) for more details.

- `url` - (Required) The Webhook URL.
- `body_template` - (Optional) The Webhook template body.

#### Zendesk Arguments

> See [the Zendesk integration documentation](https://docs.ilert.com/integrations/zendesk) for more details.

- `priority` - (Required) The Zendesk priority. Allowed values are `urgent`, `high`, `normal`, `low`.

#### GitHub Arguments

> See [the GitHub outbound issue integration documentation](https://docs.ilert.com/integrations/github/outbound-issue) for more details.

- `owner` - (Required) The GitHub organization or repo owner.
- `repository` - (Required) The GitHub repository.
- `labels` - (Optional) A list of the GitHub labels.

#### TOPdesk Arguments

> See [the TOPdesk integration documentation](https://docs.ilert.com/integrations/topdesk/outbound) for more details.

- `status` - (Required) The TOPdesk status. Allowed values are `firstLine`, `secondLine`, `partial`. Default: `firstLine`.

#### AWS Lambda Arguments

> See [the AWS Lambda integration documentation](https://docs.ilert.com/integrations/aws-lambda) for more details.

- `url` - (Required) The AWS Lambda URL.
- `body_template` - (Optional) The AWS Lambda template body.

#### Azure Function Arguments

> See [the Azure Function integration documentation](https://docs.ilert.com/integrations/azure-functions) for more details.

- `url` - (Required) The Azure Function URL.
- `body_template` - (Optional) The Azure Function template body.

#### Google Cloud Function Arguments

> See [the Google Cloud Function integration documentation](https://docs.ilert.com/integrations/gcf) for more details.

- `url` - (Required) The Google Cloud Function URL.
- `body_template` - (Optional) The Google Cloud Function template body.

#### Email Arguments

> See [the Email Outbound integration documentation](https://docs.ilert.com/integrations/email-outbound-integration) for more details.

- `recipients` - (Required) The list of the email recipients.
- `subject` - (Required) The email subject.
- `body_template` - (Optional) The email template body.

#### Sysdig Arguments

> See [the Sysdig outbound integration documentation](https://docs.ilert.com/integrations/sysdig/outbound) for more details.

- `tags` - (Optional) The list of the Sysdig tags.
- `event_filter` - (Optional) The Sysdig event filter.

#### Zapier Arguments

> See [the Zapier Outbound integration documentation](https://docs.ilert.com/integrations/zapier/outbound) for more details.

- `url` - (Required) The Zapier trigger URL.

#### Autotask Arguments

> See [the Autotask outbound integration documentation](https://docs.ilert.com/integrations/autotask/outbound) for more details.

- `queue_id` - (Required) The Autotask Queue ID.
- `company_id` - (Optional) The Autotask Company ID.
- `issue_type` - (Optional) The Autotask Issue Type.
- `ticket_category` - (Optional) The Autotask Ticket Category.
- `ticket_type` - (Optional) The Autotask Ticket Type.

#### Zammad Arguments

> See [the Zammad outbound integration documentation](https://docs.ilert.com/integrations/zammad/outbound) for more details.

- `email` - (Required) The Zammad operator email.

#### StatusPage Arguments

> See [the StatusPage outbound integration documentation](https://docs.ilert.com/integrations/statuspage) for more details.

- `page_id` - (Required) The StatusPage Page ID.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the alert action.
- `name` - The name of the alert action.
- `created_at` - The creation date time of the alert action in in ISO 8601 format.
- `updated_at` - The creation date time of the alert action in in ISO 8601 format.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_alert_action.main 123456789
```