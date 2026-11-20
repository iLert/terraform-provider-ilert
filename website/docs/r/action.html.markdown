---
layout: "ilert"
page_title: "ilert: ilert_action"
sidebar_current: "docs-ilert-resource-action"
subcategory: "Early access"
description: |-
  Creates and manages an action in ilert.
---

# ilert_action

An action calls another system when an alert or incident changes: it posts to a chat or a webhook, opens a ticket, sends an email or reroutes the alert. It is built from a template that ilert maintains, or defines its own HTTP steps. Actions are the successor to [alert actions](alert_action.html), which keep working next to them.

~> **Early access:** actions are an [early access](https://docs.ilert.com/developer-docs/rest-api/api-reference/introduction/early-access) resource of the ilert API, and disabled by default: add `"actions"` (for `/api/beta/actions`) to `early_access_enabled` in the [provider configuration](../index.html#early-access) to use this resource and its data source. The provider sends the `ilert-beta` header the API requires for them. They may change or be removed, and so may this resource, even in a minor version of the provider. ilert also enables actions account by account, and an account without them answers that actions are not available.

## Example Usage

```hcl
provider "ilert" {
  early_access_enabled = ["actions", "credentials"]
}

# a credential of type NONE holds the whole webhook URL
resource "ilert_credential" "discord" {
  name         = "discord ops"
  type         = "NONE"
  host_url     = var.discord_webhook_url
  category     = "global-action"
  sub_category = "discord"
}

resource "ilert_action" "discord" {
  name           = "post high priority alerts to discord"
  template_key   = "discord_alert"
  entity_kind    = "ALERT"
  credential_id  = ilert_credential.discord.id
  trigger_events = ["alert.created", "alert.accepted", "alert.resolved"]
  condition_icl  = "entity.source.id in [\"${ilert_alert_source.example.id}\"] and entity.priority == \"HIGH\""

  config = jsonencode({
    pushComments = true
    editMessage  = false
  })
}

resource "ilert_action" "email" {
  name          = "email the on-call lead"
  template_key  = "email_alert"
  entity_kind   = "ALERT"
  condition_icl = "entity.source.id in [\"${ilert_alert_source.example.id}\"]"

  config = jsonencode({
    recipients = ["lead@example.com"]
    subject    = "{{ entity.summary }}"
  })
}
```

### An action with its own steps

A system without a template, such as an internal API, takes the requests the action sends as `steps`:

```hcl
resource "ilert_credential" "tickets" {
  name         = "internal tickets"
  type         = "BEARER"
  host_url     = "https://tickets.example.com"
  secret_value = var.tickets_token
  category     = "global-action"
  sub_category = "custom"
}

resource "ilert_action" "tickets" {
  name          = "open a ticket"
  entity_kind   = "ALERT"
  credential_id = ilert_credential.tickets.id
  condition_icl = "entity.source.id in [\"${ilert_alert_source.example.id}\"]"

  steps = jsonencode([{
    id        = "open-ticket"
    when      = "twin == nil"
    operation = "CREATE"
    requests = [{
      id      = "post"
      method  = "POST"
      url     = "{{ credential.hostUrl }}/api/tickets"
      headers = { "Content-Type" = "application/json" }
      body    = jsonencode({ title = "{{ entity.summary }}" })
      expect  = { successStatus = [200, 201] }
      retry   = "NEVER"
    }]
    referenceWrite = {
      externalId = "{{ last.body.id }}"
    }
  }])
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the action, unique across the actions and alert actions of the account.
- `entity_kind` - (Required) What the action is bound to. Allowed values are `ALERT` and `INCIDENT`. It cannot be changed, so changing it replaces the action.
- `template_key` - (Optional) The template the action is built from, such as `discord_alert`, `email_alert` or `jira_cloud_alert`. The API lists the templates of the account at `GET /api/beta/actions/templates`, with what each needs at `GET /api/beta/actions/templates/{key}`. The API cannot change or remove the template of an action, so changing or removing it replaces the action. An action without a template needs `steps`.
- `template_version` - (Optional) Pins the version of the template. Left out, the action follows the latest version.
- `description` - (Optional) The description of the action.
- `enabled` - (Optional) Whether the action runs. Default: `true`.
- `dispatch_mode` - (Optional) `AUTOMATIC` runs the action on every matching event, `MANUAL` only when it is run on an alert or incident in ilert. It cannot be changed, so changing it replaces the action. Default: `AUTOMATIC`.
- `trigger_events` - (Optional) The events the action reacts to. Left out, it reacts to every event of its entity kind. `alert.not_resolved` and `incident.not_resolved` need a `trigger_event_delay`. An `ALERT` action takes `alert.created`, `alert.accepted`, `alert.resolved`, `alert.rerouted`, `alert.escalated`, `alert.escalation_ended`, `alert.priority_raised`, `alert.responder_added`, `alert.responder_removed`, `alert.comment_added`, `alert.channel_attached`, `alert.channel_detached` and `alert.not_resolved`. An `INCIDENT` action takes `incident.opened`, `incident.updated`, `incident.status_changed`, `incident.resolved`, `incident.deleted`, `incident.responder_added`, `incident.responder_removed`, `incident.responder_updated`, `incident.comment_added`, `incident.channel_attached`, `incident.channel_detached`, `incident.postmortem_linked` and `incident.not_resolved`.
- `trigger_event_delay` - (Optional) One or more [trigger event delay](#trigger-event-delay-arguments) blocks, at most one per event.
- `condition_icl` - (Optional) The condition an alert or incident has to meet for the action to run, in the [ilert condition language](https://docs.ilert.com/rest-api/icl-ilert-condition-language). Left out, the condition of the template applies. The condition also decides whose alerts or incidents the action acts on, see [Condition and permissions](#condition-and-permissions).
- `config` - (Optional) The values for the fields the template declares in its `configSpec`, as a JSON object written with `jsonencode`. Each value takes the type of its field: a `checkbox` a boolean, `chips` a list, a `number` a number. The defaults of a template are not applied when the action runs, so set the values it needs. Only the keys the configuration sets are managed, see [What the provider leaves alone](#what-the-provider-leaves-alone).
- `credential_id` - (Optional) The ID of the [credential](credential.html) the action authenticates with. The kinds of credential a template takes are its `authSpec`. Like the web app, give a credential for actions the `category` `global-action` and the brand of the template as `sub_category`. Left out, the action has no credential.
- `steps` - (Optional) The requests the action sends, as a JSON array written with `jsonencode`, see [the actions documentation of ilert](https://docs.ilert.com/integrations/actions). An action built from a template runs the template's steps while this is left out, and your steps replace them completely. Leave `steps`, `field_mapping` and `inbound` out to use a template as it is, which also picks up the changes ilert makes to it: a copy of the template's own definition would no longer follow them.
- `field_mapping` - (Optional) The complete list of field mappings, the template's with your changes applied, as a JSON array written with `jsonencode`.
- `inbound` - (Optional) How a bidirectional action turns a callback into a change in ilert, as a JSON object written with `jsonencode`. It replaces the template's handling as a whole.
- `bidirectional` - (Optional) Whether the other system may send changes back to ilert, through `integration_url`. Requires a template that supports it. Default: `false`.
- `team` - (Optional) One or more [team](#team-arguments) blocks. The order in which the blocks are declared is not significant.
- `labels` - (Optional) Labels of the action, as a map of strings.

#### Trigger Event Delay Arguments

- `event` - (Required) The event to delay: `alert.not_resolved` and `incident.not_resolved`, which fire once the alert or incident is still open after the delay, or `alert.escalation_ended`.
- `delay_sec` - (Required) The delay in seconds, 60 to 172800 for `alert.not_resolved` and `incident.not_resolved`, and 30 to 7200 for `alert.escalation_ended`.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

### Condition and permissions

ilert checks the condition against the user of the API token. A user who is not an admin has to restrict every `or` branch of it to alerts or incidents they can reach: for alerts their teams (`entity.teamIds`) or alert sources they can read (`entity.source.id`), for incidents their teams or services they can read (`entity.serviceIds`), with the IDs as strings, such as `entity.source.id in ["1234"]`.

An action without a condition, whose template does not set one either, runs on every alert or every incident of the account, those of alert sources and services created later included. Only one enabled `AUTOMATIC` action per entity kind may run without a condition.

### What the provider leaves alone

Some templates write values into the config of an action when it is created or updated. The API replaces the config with what an update sends, so the provider reads the action before every update and sends its config back with the configured keys applied, and a key removed from `config` taken out. The values ilert wrote stay in ilert, and the state only holds the keys the configuration sets, so they do not show as a change.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the action.
- `name` - The name of the action.
- `integration_url` - Where the other system sends the callbacks of a bidirectional action. It contains the key that authenticates them, so keep it out of logs and tickets.
- `brand` - The vendor the action talks to, such as `discord`, taken from its template. Empty for an action with its own steps.
- `status` - The outcome of the most recent run of the action: `NEVER_TRIGGERED`, `OK` or `ERROR`.
- `created_at` - The creation date of the action in ISO format.
- `updated_at` - The date the action was last updated, in ISO format.

## Import

Actions can be imported using the `id`, e.g.

```sh
$ terraform import ilert_action.main 123456789
```

`config` is not captured by the import: its keys are managed once the configuration sets them, and the import has no configuration to read them from. The first plan after importing applies the keys the configuration sets, and leaves the others as they are, the values ilert wrote included.

The API lists the alert actions of the account among the actions, as mirrors that cannot be changed. The provider refuses to import or manage one, deleting it would delete the alert action: manage those with [`ilert_alert_action`](alert_action.html).
