---
layout: "ilert"
page_title: "ilert: ilert_alert_source"
sidebar_current: "docs-ilert-resource-alert-source"
description: |-
  Creates and manages an alert source in ilert.
---

# ilert_alert_source

An [alert source](https://docs.ilert.com/getting-started/readme#alert-source-aka-inbound-integration) represents the connection between your tools (usually a monitoring system, a ticketing tool, or an application) and ilert. We often refer to alert sources as inbound integrations.

## Example Usage

```hcl
resource "ilert_user" "example" {
  email      = "example@example.com"
  first_name = "example"
  last_name  = "example"
}

resource "ilert_escalation_policy" "example" {
  name = "example"

  escalation_rule {
    escalation_timeout = 15
    user               = ilert_user.example.id
  }
}

resource "ilert_alert_source" "example" {
  name              = "My Grafana Integration from terraform"
  integration_type  = "GRAFANA"
  escalation_policy = ilert_escalation_policy.example.id
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the alert source.
- `integration_type` - (Required) The integration type of the alert source. Allowed values are `NAGIOS`, `ICINGA`, `EMAIL2`, `SMS`, `API`, `CRN`, `PRTG`, `PINGDOM`, `CLOUDWATCH`, `AWSPHD`, `STACKDRIVER`, `INSTANA`, `ZABBIX`, `SOLARWINDS`, `PROMETHEUS`, `NEWRELIC`, `GRAFANA`, `GITHUB`, `DATADOG`, `UPTIMEROBOT`, `APPDYNAMICS`, `DYNATRACE`, `TOPDESK`, `STATUSCAKE`, `MONITOR`, `TOOL`, `CHECKMK`, `AUTOTASK`, `AWSBUDGET`, `KENTIXAM`, `JIRA`, `CONSUL`, `ZAMMAD`, `SIGNALFX`, `SPLUNK`, `KUBERNETES`, `SEMATEXT`, `SENTRY`, `SUMOLOGIC`, `RAYGUN`, `MXTOOLBOX`, `ESWATCHER`, `AMAZONSNS`, `KAPACITOR`, `CORTEXXSOAR`, `SYSDIG`, `SERVERDENSITY`, `ZAPIER`, `SERVICENOW`, `SEARCHGUARD`, `AZUREALERTS`, `TERRAFORMCLOUD`, `ZENDESK`, `AUVIK`, `SENSU`, `NCENTRAL`, `JUMPCLOUD`, `SALESFORCE`, `GUARDDUTY`, `STATUSHUB`, `IXON`, `APIFORTRESS`, `FRESHSERVICE`, `APPSIGNAL`, `LIGHTSTEP`, `IBMCLOUDFUNCTIONS`, `CROWDSTRIKE`, `HUMIO`, `OHDEAR`, `MONGODBATLAS`, `GITLAB`, `HYPERPING`, `PAPRISMA`, `SAMSARA`, `PANDORAFMS`, `MSSCOM`, `TWILIO`, `CISCOMERAKI`, `CHECKLY`, `POSTHOG`, `GOOGLESCC`, `SLACK`, `MSTEAMS`, `UPTIMEKUMA`, `TWILIOERRORS`, `PARTICLE`, `CLOUDFLARE`, `TULIP`, `GRAYLOG`, `CATCHPOINT`, `LOKI`, `CORTEX`, `MIMIR`, `HALOPSA`, `INFLUXDB`, `CALLFLOW`, `HALOITSM`, `KIBANA`, `VICTORIAMETRICS`, `HONEYCOMB`, `FOURME`, `KEEP`, `UBIDOTS`, `HETRIXTOOLS`, `POSTMAN`, `CLUSTERCONTROL`, `NETDATA`, `AWX`, `KAFKA`, `MQTT`, `RAPIDSPIKE`, `HONEYBADGER`, `HEALTHCHECKSIO`, `MEZMO`, `SERVERGUARD24`, `CISCOTHOUSANDEYES`, `SITE24X7`, `ITCONDUCTOR`, `SAPFRUN`, `APICA`, `DASH0`, `ROLLBAR`, `GATUS`, `LIBRENMS`, `PANTHER`, `TEAMCITY`, `ALIBABACLOUD`, `FLEETDM`, `CONNECTWISEPSA`, `DEADMANSSNITCH`, `FORTISOAR`, `OPMANAGER`, `CRONITOR`, `DOMOTZ`, `LIVEWATCH`, `AZUREDEVOPS`, `LEVELIO`, `EKARA`, `SYSAID`, `PHAREIO`, `OPSGENIE`, `WHATAP`, `SIGNOZ`, `HEARTBEAT2`. If an integration type is missing, please check: https://docs.ilert.com/developer-docs/rest-api/api-reference/alert-sources#post-alert-sources.
- `escalation_policy` - (Required) The escalation policy id used by this alert source.
- `alert_creation` - (Optional) ilert receives events from your monitoring systems and can then create alerts in different ways. This option is recommended. Allowed values are `ONE_ALERT_PER_EMAIL`, `ONE_ALERT_PER_EMAIL_SUBJECT`, `ONE_PENDING_ALERT_ALLOWED`, `ONE_OPEN_ALERT_ALLOWED`, `OPEN_RESOLVE_ON_EXTRACTION`, `ONE_ALERT_GROUPED_PER_WINDOW`, `INTELLIGENT_GROUPING`. `alert_grouping_window` must be defined when this field is set to `ONE_ALERT_GROUPED_PER_WINDOW` or `INTELLIGENT_GROUPING`.
- `active` - (Optional) The state of the alert source. Default: `true`.
- `alert_priority_rule` - (Optional) The alert priority rule. This option is recommended. Allowed values are `HIGH`, `LOW`, `HIGH_DURING_SUPPORT_HOURS`, `LOW_DURING_SUPPORT_HOURS`.
- `auto_resolution_timeout` - (Optional) The auto resolution timeout. Allowed values are `PT10M`, `PT20M`, `PT30M`, `PT40M`, `PT50M`, `PT60M`, `PT90M`, `PT2H`, `PT3H`, `PT4H`, `PT5H`, `PT6H`, `PT12H`, `PT24H` (`H` means hour and `M` means minute).
- `email` - (Optional) The email address. This option is required if `integration_type` is `EMAIL2`.
- `support_hours` - (Optional) A [support_hours](#support-hours-arguments) block. This option is allowed if `alert_priority_rule` is `HIGH_DURING_SUPPORT_HOURS` or `LOW_DURING_SUPPORT_HOURS`.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `summary_template` - (Optional) A summary [template](#template-arguments) block.
- `details_template` - (Optional) A details [template](#template-arguments) block.
- `routing_template` - (Optional) A routing [template](#template-arguments) block.
- `alert_key_template` - (Optional) An alert key [template](#template-arguments) block.
- `link_template` - (Optional) One or more [link template](#link-template-arguments) block.
- `priority_template` - (Optional) A [priority template](#priority-template-arguments) block.
- `alert_grouping_window` - (Optional) The alert grouping time frame. Any alerts triggered within this time frame will be grouped together. This field has to be defined when `alert_creation` is set to `ONE_ALERT_GROUPED_PER_WINDOW` or `INTELLIGENT_GROUPING`.
- `score_threshold` - (Optional) Sets the score threshold. Indicates how similar alerts are to be grouped together. This field has to be defined when `alert_creation` is set to `INTELLIGENT_GROUPING`. Should be a floating point number between `0` and `1` (inclusive).
- `event_filter` - (Optional) Defines event filter condition in ICL language. This is a code based implementation, more info on syntax: https://docs.ilert.com/rest-api/icl-ilert-condition-language. For block based configuration please use the web UI.
- `event_type_filter_create` - (Optional) Defines an event type filter condition for create events in ICL language. This is a code based implementation, more info on syntax: https://docs.ilert.com/rest-api/icl-ilert-condition-language. For block based configuration please use the web UI.
- `event_type_filter_accept` - (Optional) Defines an event type filter condition for accept events in ICL language. This is a code based implementation, more info on syntax: https://docs.ilert.com/rest-api/icl-ilert-condition-language. For block based configuration please use the web UI.
- `event_type_filter_resolve` - (Optional) Defines an event type filter condition for resolve events in ICL language. This is a code based implementation, more info on syntax: https://docs.ilert.com/rest-api/icl-ilert-condition-language. For block based configuration please use the web UI.

#### Support Hours Arguments

- `id` - The id of the support hour given as reference.

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

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### Template Arguments

- `text_template` - (Required) The content of the template. It is recommended to use the exact content as generated via blocks in the web UI to prevent inconsistencies between the ilert API and Terraform.

#### Link template Arguments

- `text` - (Required) The display name for the link.
- `href_template` - (Required) A [template](#template-arguments) block.

#### Priority template Arguments

- `value_template` - (Required) A [template](#template-arguments) block.
- `mapping` - (Required) One or more [mapping](#mapping-arguments) blocks.

#### Mapping Arguments

- `value` - (Required) The value that should be extracted from the alerts payload.
- `priority` - (Required) The priority the alert should be mapped to. Allowed values are `HIGH` and `LOW`.

### Support Hours Example

```hcl
resource "ilert_support_hour" "example" {
  name = "example"
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

resource "ilert_alert_source" "example_with_support_hours" {
  name                = "My Grafana Integration from terraform with support hours"
  integration_type    = "GRAFANA"
  escalation_policy   = ilert_escalation_policy.example.id
  alert_priority_rule = "HIGH_DURING_SUPPORT_HOURS"

  support_hours {
    id = ilert_support_hour.example.id
  }
}
```

### Email example

```hcl
resource "ilert_alert_source" "example_email" {
  name              = "My Email Integration from terraform"
  integration_type  = "EMAIL"
  email             = "example@ 'your tenant' .ilert.eu"
  escalation_policy = ilert_escalation_policy.example.id

  alert_creation = "OPEN_RESOLVE_ON_EXTRACTION"
  resolve_key_extractor {
    field    = "EMAIL_SUBJECT"
    criteria = "ALL_TEXT_BEFORE"
    value    = "my server"
  }

  email_filtered = true
  email_predicate {
    field    = "EMAIL_BODY"
    criteria = "CONTAINS_STRING"
    value    = "alarm"
  }

  email_resolve_filtered = true
  email_resolve_predicate {
    field    = "EMAIL_BODY"
    criteria = "CONTAINS_STRING"
    value    = "resolve"
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

Alert sources can be imported using the `id`, e.g.

```sh
$ terraform import ilert_alert_source.main 123456789
```
