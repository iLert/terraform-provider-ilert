---
layout: "ilert"
page_title: "iLert: ilert_alert_source"
sidebar_current: "docs-ilert-resource-alert-source"
description: |-
  Creates and manages an alert source in iLert.
---

# ilert_alert_source

An [alert source](https://api.ilert.com/api-docs/#tag/Alert-Sources) represents the connection between your tools (usually a monitoring system, a ticketing tool, or an application) and iLert. We often refer to alert sources as inbound integrations.

## Example Usage

```hcl
data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_alert_source" "example" {
  name                    = "My Grafana Integration"
  integration_type        = "GRAFANA"
  escalation_policy       = data.ilert_escalation_policy.default.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the alert source.
- `integration_type` - (Required) The integration type of the alert source. Allowed values are `NAGIOS`, `ICINGA`, `EMAIL`, `SMS`, `API`, `CRN`, `HEARTBEAT`, `PRTG`, `PINGDOM`, `CLOUDWATCH`, `AWSPHD`, `STACKDRIVER`, `INSTANA`, `ZABBIX`, `SOLARWINDS`, `PROMETHEUS`, `NEWRELIC`, `GRAFANA`, `GITHUB`, `DATADOG`, `UPTIMEROBOT`, `APPDYNAMICS`, `DYNATRACE`, `TOPDESK`, `STATUSCAKE`, `MONITOR`, `TOOL`, `CHECKMK`, `AUTOTASK`, `AWSBUDGET`, `KENTIXAM`, `JIRA`, `CONSUL`, `ZAMMAD`, `SIGNALFX`, `SPLUNK`, `KUBERNETES`, `SEMATEXT`, `SENTRY`, `SUMOLOGIC`, `RAYGUN`, `MXTOOLBOX`, `ESWATCHER`, `AMAZONSNS`, `KAPACITOR`, `CORTEXXSOAR`.
- `escalation_policy` - (Required) The escalation policy id used by this alert source.
- `incident_creation` - (Optional) iLert receives events from your monitoring systems and can then create incidents in different ways. This option is recommended. Allowed values are `ONE_INCIDENT_PER_EMAIL`, `ONE_INCIDENT_PER_EMAIL_SUBJECT`, `ONE_PENDING_INCIDENT_ALLOWED`, `ONE_OPEN_INCIDENT_ALLOWED`, `OPEN_RESOLVE_ON_EXTRACTION`.
- `active` - (Optional) The state of the alert source. Default: `true`.
- `incident_priority_rule` - (Optional) The incident priority rule. This option is recommended. Allowed values are `HIGH`, `LOW`, `HIGH_DURING_SUPPORT_HOURS`, `LOW_DURING_SUPPORT_HOURS`.
- `auto_resolution_timeout` - (Optional) The auto resolution timeout. Allowed values are `PT10M`, `PT20M`, `PT30M`, `PT40M`, `PT50M`, `PT60M`, `PT90M`, `PT2H`, `PT3H`, `PT4H`, `PT5H`, `PT6H`, `PT12H`, `PT24H` (`H` means hour and `M` means minute).
- `email` - (Optional) The email address. This option is required if `integration_type` is `EMAIL`.
- `email_filtered` - (Optional) The email filtered. This option is required if `integration_type` is `EMAIL`.
- `email_resolve_filtered` - (Optional) The email resolve filtered. This option is required if `integration_type` is `EMAIL`.
- `filter_operator` - (Optional) The filter operator. This option is required if `integration_type` is `EMAIL`. Allowed values are `AND` and `OR`.
- `resolve_filter_operator` - (Optional) The resolve filter operator. This option is required if `integration_type` is `EMAIL`. Allowed values are `AND` and `OR`.
- `resolve_key_extractor` - (Optional) A [resolve key extractor](#resolve-key-extractor-arguments) block. This option is required if `integration_type` is `EMAIL`.
- `email_predicate` - (Optional) One or more [email predicate](#email-predicate-arguments) blocks. This option is required if `integration_type` is `EMAIL`.
- `email_resolve_predicate` - (Optional) One or more [email resolve predicate](#email-resolve-predicate-arguments) blocks. This option is required if `integration_type` is `EMAIL`.
- `heartbeat` - (Optional) A [heartbeat](#heartbeat-arguments) block. This option is required if `integration_type` is `HEARTBEAT`.
- `support_hours` - (Optional) A [support_hours](#support-hours-arguments) block. This option is allowed if `incident_priority_rule` is `HIGH_DURING_SUPPORT_HOURS` or `LOW_DURING_SUPPORT_HOURS`.
- `autotask_metadata` - (Optional) An [autotask metadata](#autotask-metadata-arguments) block. This option is required if `integration_type` is `AUTOTASK`.
- `teams` - (Optional) A list of related team ids.

#### Heartbeat Arguments

- `summary` - The summary of the heartbeat.
- `interval_sec` - The interval in seconds of the heartbeat. Default: `900`

#### Support Hours Arguments

- `timezone` - The timezone of the support hours.
- `auto_raise_incidents` - Raise priority of all pending incidents for this alert source to 'high' when support hours begin.
- `support_days` - The [support days](#support-days-arguments) block of the support hours.

#### Support Days Arguments

- `monday` - The [support day](#support-day-arguments) block of the support days.
- `tuesday` - The [support day](#support-day-arguments) block of the support days.
- `wednesday` - The [support day](#support-day-arguments) block of the support days.
- `thursday` - The [support day](#support-day-arguments) block of the support days.
- `friday` - The [support day](#support-day-arguments) block of the support days.
- `saturday` - The [support day](#support-day-arguments) block of the support days.
- `sunday` - The [support day](#support-day-arguments) block of the support days.

#### Support Day Arguments

- `start` - The start time of the support day.
- `end` - The end time of the support day.

#### Autotask Metadata Arguments

- `username` - The username of the autotask api user.
- `secret` - The secret of the autotask api user.
- `web_server` - The Autotask API server URL. Default: `https://webservices2.autotask.net`

#### Resolve Key Extractor Arguments

- `field` - The field of the resolve key extractor. Allowed values are `EMAIL_SUBJECT` and `EMAIL_BODY`.
- `criteria` - The criteria of the resolve key extractor. Allowed values are `ALL_TEXT_BEFORE`, `ALL_TEXT_AFTER` and `MATCHES_REGEX`.
- `value` - The value of the resolve key extractor.

#### Email Predicate Arguments

- `field` - The field of the email predicate. Allowed values are `EMAIL_FROM`, `EMAIL_SUBJECT` and `EMAIL_BODY`.
- `criteria` - The criteria of the email predicate. Allowed values are `CONTAINS_ANY_WORDS`, `CONTAINS_NOT_WORDS`, `CONTAINS_STRING`, `CONTAINS_NOT_STRING`, `IS_STRING`, `IS_NOT_STRING`, `MATCHES_REGEX`, `MATCHES_NOT_REGEX`.
- `value` - The value of the email predicate.

#### Email Resolve Predicate Arguments

- `field` - The field of the email resolve predicate. Allowed values are `EMAIL_FROM`, `EMAIL_SUBJECT` and `EMAIL_BODY`.
- `criteria` - The criteria of the email resolve predicate. Allowed values are `CONTAINS_ANY_WORDS`, `CONTAINS_NOT_WORDS`, `CONTAINS_STRING`, `CONTAINS_NOT_STRING`, `IS_STRING`, `IS_NOT_STRING`, `MATCHES_REGEX`, `MATCHES_NOT_REGEX`.
- `value` - The value of the email resolve predicate.

### Support Hours Example

```hcl
data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_alert_source" "example" {
  name                   = "My Grafana Integration"
  integration_type       = "GRAFANA"
  escalation_policy      = data.ilert_escalation_policy.default.id
  incident_priority_rule = "HIGH_DURING_SUPPORT_HOURS"

  support_hours {
    timezone = "Europe/Berlin"

    support_days {
      monday {
        start = "08:00"
        end   = "17:00"
      }

      tuesday {
        start = "08:00"
        end   = "17:00"
      }

      wednesday {
        start = "08:00"
        end   = "17:00"
      }

      thursday {
        start = "08:00"
        end   = "17:00"
      }

      friday {
        start = "08:00"
        end   = "17:00"
      }
    }
  }
}
```

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the alert source.
- `name` - The name of the found alert source.
- `status` - The status of the found alert source.
- `integration_key` - The integration key of the found alert source.
- `integration_url` - The integration URL of the found alert source.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_alert_source.main 123456789
```
