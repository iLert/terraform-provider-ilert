---
layout: "ilert"
page_title: "ilert: ilert_alert_action"
sidebar_current: "docs-ilert-resource-alert-action"
description: |-
  Creates and manages an alert action in ilert.
---

# ilert_alert_action

An [alert_action](https://docs.ilert.com/getting-started/readme#connectors-and-alert-actions-aka-outbound-integrations) is created at the alert source level and uses its [connector](connector.html) to perform a specified action.

## Example Usage

```hcl
resource "ilert_user" "example" {
  first_name = "example"
  last_name  = "example"
  email      = "example@example.com"
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
  escalation_policy = ilert_escalation_policy.example.id
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
- `alert_source` - (Required) One or more [alert source](#alert-source-arguments) blocks.
- `connector` - (Required) A [connector](#connector-arguments) block.
- `trigger_mode` - (Optional) The trigger mode of the alert action. Allowed values are `AUTOMATIC` or `MANUAL`. Default: `AUTOMATIC`.
- `escalation_ended_delay_sec` - (Optional) The number of seconds the alert action will be delayed when reaching end of escalation. Should only be set when one of `trigger_types` is set to `alert-escalation-ended`. Must be either `0` or a value between `30` and `7200`.
- `not_resolved_delay_sec` - (Optional) The number of seconds the alert action will be delayed when the alert is not resolved yet. Should only be set when one of `trigger_types` is set to `v-alert-not-resolved`. Must be either `0` or a value between `60` and `7200`.
- `trigger_types` - (Optional if the `MANUAL` trigger mode and required if the `AUTOMATIC` trigger mode) A list of the trigger types. Allowed values are `alert-created`, `alert-assigned`, `alert-auto-escalated`, `alert-acknowledged`, `alert-raised`, `alert-comment-added`, `alert-escalation-ended`, `alert-resolved`, `alert-auto-resolved`, `alert-responder-added`, `alert-responder-removed`, `alert-channel-attached`, `alert-channel-detached`, `v-alert-not-resolved`.
- `jira` - (Optional) A [jira](#jira-arguments) block.
- `servicenow` - (Optional) A [servicenow](#servicenow-arguments) block.
- `slack` - (Optional) A [slack](#slack-arguments) block.
- `webhook` - (Optional) A [webhook](#webhook-arguments) block.
- `zendesk` - (Optional) A [zendesk](#zendesk-arguments) block.
- `github` - (Optional) A [github](#github-arguments) block.
- `topdesk` - (Optional) A [topdesk](#topdesk-arguments) block.
- `email` - (Optional) A [email](#email-arguments) block.
- `autotask` - (Optional) A [autotask](#autotask-arguments) block.
- `zammad` - (Optional) A [zammad](#zammad-arguments) block.
- `dingtalk` - (Optional) A [dingtalk](#dingtalk-arguments) block.
- `dingtalk_action` - (Optional) A [dingtalk_action](#dingtalk-action-arguments) block.
- `automation_rule` - (Optional) An [automation_rule](#automation-rule-arguments) block.
- `telegram` - (Optional) An [telegram](#telegram-arguments) block.
- `microsoft_teams_bot` - (Optional) A [microsoft_teams_bot](#microsoft-teams-bot-arguments) block.
- `microsoft_teams_webhook` - (Optional) A [microsoft_teams_webhook](#microsoft-teams-webhook-arguments) block.
- `slack_webhook` - (Optional) A [slack_webhook](#slack-webhook-arguments) block.
- `alert_filter` - (Optional) An [alert_filter](#alert-filter-arguments) block.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `conditions` - (Optional) Defines event filter condition in ICL language. This is a code based implementation, more info on syntax: https://docs.ilert.com/rest-api/icl-ilert-condition-language. For block based configuration please use the web UI.

#### Alert Source Arguments

- `id` - (Required) The alert source id.

#### Connector Arguments

- `id` - (Optional) The connector id. Required if the connector `type` is one of values `jira`, `servicenow`, `zendesk`, `discord`, `github`, `topdesk`, `autotask`, `mattermost`, `zammad`, `dingtalk`, `microsoft_teams_bot`, `slack`.
- `type` - (Required) The connector type. Allowed values are `jira`, `servicenow`, `zendesk`, `discord`, `github`, `topdesk`, `autotask`, `mattermost`, `zammad`, `dingtalk`, `microsoft_teams_bot`, `slack`.

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
- `body_template` - (Optional) The custom template body.

#### Slack Arguments

> See [the Slack integration documentation](https://docs.ilert.com/integrations/slack) for more details.

- `channel_id` - (Required) The Slack channel id. Unique value.
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

#### Email Arguments

> See [the Email Outbound integration documentation](https://docs.ilert.com/integrations/email-outbound-integration) for more details.

- `recipients` - (Required) The list of the email recipients.
- `subject` - (Required) The email subject.
- `body_template` - (Optional) The email template body.

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

#### Dingtalk Arguments

> See [the Dingtalk outbound integration documentation](https://docs.ilert.com/integrations/dingtalk) for more details.

- `is_at_all` - (Optional) Determines whether messages are sent with `@all` or not. Allowed values are `true` or `false`.
- `at_mobiles` - (Optional) Mobile numbers to notify related DingTalk users.

#### Dingtalk action Arguments

> See [the Dingtalk action outbound integration documentation](https://docs.ilert.com/integrations/dingtalk) for more details.

- `url` - (Required) The Dingtalk action URL.
- `secret` - (Optional) The secret for the provided Dingtalk URL.
- `is_at_all` - (Optional) Determines whether messages are sent with `@all` or not. Allowed values are `true` or `false`.
- `at_mobiles` - (Optional) Mobile numbers to notify related DingTalk users.

#### Automation Rule Arguments

- `alert_type` - (Required) The alert type. Allowed values are `CREATED` or `ACCEPTED`.
- `service_ids` - (Required) One or more service ID's.
- `service_status` - (Required) The status the service should be set in. Allowed values are `OPERATIONAL`, `UNDER_MAINTENANCE`, `DEGRADED`, `PARTIAL_OUTAGE`, `MAJOR_OUTAGE`.
- `template_id` - (Optional) The ID of the incident template.
- `resolve_incident` - (Optional, requires `template_id`) Determines whether an incident should be resolved or not. Default: `false`
- `send_notification` - (Optional, requires `template_id`) Determines whether notifications should be sent or not. Default: `false`

#### Telegram Arguments

> See [the Telegram integration documentation](https://docs.ilert.com/integrations/telegram) for more details.

- `channel_id` - (Required) The Telegram channel id.

#### Microsoft teams bot Arguments

> See [the Microsoft teams bot integration documentation](https://docs.ilert.com/chatops/microsoft-teams) for more details.

- `channel_id` - (Required) The id of the channel.
- `channel_name` - (Optional) The name of the channel.
- `team_id` - (Required) The id of the team.
- `team_name` - (Optional) The name of the team.
- `type` - (Required) The type of the bot setup. Allowed values are `chat` or `meeting`.

#### Microsoft teams webhook Arguments

- `url` - (Required) The workflow URL for the channel.
- `body_template` - (Optional) The custom template body.

#### Slack webhook Arguments

- `url` - (Required) The workflow URL for the channel.

#### Alert Filter Arguments

- `operator` - (Required) The operator to use for the filter. Allowed values are `AND` or `OR`.
- `predicate` - (Required) One or more [predicate](#predicate-arguments) blocks.

#### Predicate Arguments

- `field` - (Required) The field which should be monitored for conditional execution. Allowed values are `ALERT_SUMMARY`, `ALERT_DETAILS`, `ESCALATION_POLICY`, `ALERT_PRIORITY`.
- `criteria` - (Required) The criteria for the condition. Allowed values are `CONTAINS_ANY_WORDS`, `CONTAINS_NOT_WORDS`, `CONTAINS_STRING`, `CONTAINS_NOT_STRING`, `IS_STRING`, `IS_NOT_STRING`, `MATCHES_REGEX`, `MATCHES_NOT_REGEX`.
- `value` - (Required) The value for the condition.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the alert action.
- `name` - The name of the alert action.
- `created_at` - The creation date time of the alert action in in ISO 8601 format.
- `updated_at` - The creation date time of the alert action in in ISO 8601 format.

## Import

Alert actions can be imported using the `id`, e.g.

```sh
$ terraform import ilert_alert_action.main 123456789
```
