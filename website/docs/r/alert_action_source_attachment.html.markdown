---
layout: "ilert"
page_title: "ilert: ilert_alert_action_source_attachment"
sidebar_current: "docs-ilert-resource-alert-action-source-attachment"
description: |-
  Manages a single alert-source attachment on a shared alert action without overwriting other attachments.
---

# ilert_alert_action_source_attachment

Manages **one** attachment between an alert action and an alert source. Each Terraform resource owns exactly one tuple `(alert_action_id, alert_source_id)`.

Use this resource instead of the [`alert_source` block on `ilert_alert_action`](alert_action.html) when:

- multiple teams or modules need to attach **their own** alert sources to a **shared** alert action, and
- no single owner should manage the full list of attached sources.

The resource calls the iLert `PUT /alert-actions/{id}/add-source` and `PUT /alert-actions/{id}/remove-source` endpoints, which mutate only the source you specify and leave every other attached source untouched.

~> **Do not mix** this resource with the `alert_source` block on the same `ilert_alert_action`. The `alert_source` block is authoritative — it rewrites the full source list on every apply and will fight with `ilert_alert_action_source_attachment`, producing perpetual diffs. The recommended pattern is to bootstrap the alert action with a single source via `alert_source` and then `lifecycle { ignore_changes = [alert_source] }` so subsequent attachments are managed exclusively through `ilert_alert_action_source_attachment`.

~> **Bootstrap source required.** The iLert API rejects alert action creation with no sources (`Either field 'alertSourceIds [long]' or 'alertSources [{id}]' is required`). Provide at least one initial `alert_source` block on `ilert_alert_action`, then attach additional sources via this resource.

Create and destroy are idempotent against out-of-band drift: re-attach of an already-attached source and re-detach of an already-detached source are both treated as success.

~> **Concurrency.** The iLert add-source / remove-source endpoints are read-modify-write on the alert action, so the provider serializes attach/detach operations per `alert_action_id` to keep parallel `terraform apply` (default `-parallelism=10`) from losing attachments. This guard is per Terraform process; running two separate applies against the same alert action at the same time can still race — avoid that.

## Example Usage

```hcl
resource "ilert_alert_source" "bootstrap" {
  name              = "bootstrap-source"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.example.id
}

resource "ilert_alert_source" "team_a" {
  name              = "team-a-source"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.team_a.id
}

resource "ilert_alert_source" "team_b" {
  name              = "team-b-source"
  integration_type  = "API"
  escalation_policy = ilert_escalation_policy.team_b.id
}

resource "ilert_alert_action" "shared" {
  name         = "shared-slack-action"
  trigger_mode = "AUTOMATIC"

  alert_source {
    id = ilert_alert_source.bootstrap.id
  }

  connector {
    id   = ilert_connector.slack.id
    type = "SLACK"
  }

  slack {
    channel_id   = "C0123456789"
    channel_name = "incidents"
  }

  lifecycle {
    ignore_changes = [alert_source]
  }
}

resource "ilert_alert_action_source_attachment" "team_a_to_shared" {
  alert_action {
    id = ilert_alert_action.shared.id
  }
  alert_source {
    id = ilert_alert_source.team_a.id
  }
}

resource "ilert_alert_action_source_attachment" "team_b_to_shared" {
  alert_action {
    id = ilert_alert_action.shared.id
  }
  alert_source {
    id = ilert_alert_source.team_b.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `alert_action` - (Required, ForceNew) The alert action the source is attached to. Structure [documented below](#alert_action).
- `alert_source` - (Required, ForceNew) The alert source being attached. Structure [documented below](#alert_source).

Changing either block forces a new resource.

### alert_action

- `id` - (Required, ForceNew) The ID of the alert action.

### alert_source

- `id` - (Required, ForceNew) The ID of the alert source (numeric).

## Attributes Reference

The following attributes are exported:

- `id` - Composite ID in the form `<alert_action_id>/<alert_source_id>`.

## Import

Use the composite ID `<alert_action_id>/<alert_source_id>`:

```sh
$ terraform import ilert_alert_action_source_attachment.main 123456789/987654321
```
