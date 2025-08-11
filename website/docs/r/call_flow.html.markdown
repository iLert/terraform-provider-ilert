---
layout: "ilert"
page_title: "ilert: ilert_call_flow"
sidebar_current: "docs-ilert-resource-call-flow"
description: |-
    Creates and manages a call flow in ilert.
---

# ilert_call_flow

A call flow defines the IVR (interactive voice response) structure for incoming calls.

## Example Usage

```hcl
resource "ilert_call_flow" "example" {
  name     = "example-call-flow"
  language = "English"

  root_node = [{
    node_type = "ROOT"
    name      = "root"

    branches = [
      {
        branch_type = "ANSWERED"
        target = [{
          node_type = "AUDIO_MESSAGE"
          name      = "welcome"
          metadata = [{
            text_message = "Welcome to our hotline."
          }]
        }]
      },
      {
        branch_type = "CATCH_ALL"
        target = [{
          node_type = "VOICEMAIL"
          name      = "vm"
          metadata = [{
            text_message = "Please leave a message after the beep."
          }]
        }]
      }
    ]
  }]
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the call flow.
- `language` - (Required) The language used by the call flow. Allowed values: `German`, `English`.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `root_node` - (Required) A single [node](#node-arguments) block defining the root of the call flow.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### Node Arguments

- `name` - (Optional) The name of the node.
- `node_type` - (Required) The node type. Allowed values: `ROOT`, `IVR_MENU`, `AUDIO_MESSAGE`, `PLAIN`, `SUPPORT_HOURS`, `ROUTE_CALL`, `VOICEMAIL`, `PIN_CODE`, `CREATE_ALERT`, `BLOCK_NUMBERS`, `AGENTIC`.
- `metadata` - (Optional) A single [metadata](#metadata-arguments) block.
- `branches` - (Optional) A list of [branch](#branch-arguments) blocks.

#### Metadata Arguments

- `text_message` - (Optional) Used by node types: `IVR_MENU`, `AUDIO_MESSAGE`, `VOICEMAIL`, `PIN_CODE`.
- `custom_audio_url` - (Optional) Used by node types: `IVR_MENU`, `AUDIO_MESSAGE`, `VOICEMAIL`, `PIN_CODE`.
- `ai_voice_model` - (Optional) Used by node types: `IVR_MENU`, `AUDIO_MESSAGE`, `VOICEMAIL`, `PIN_CODE`. Allowed values: `emma`, `liam`, `oliver`, `andreas`, `sophie`, `isabelle`, `gordon`, `bruce`, `alfred`, `ellen`, `barbara`.
- `enabled_options` - (Optional) Used by node type: `IVR_MENU`.
- `language` - (Optional) Used by node types: `IVR_MENU`, `AUDIO_MESSAGE`. Allowed values: `en`, `de`, `fr`, `es`, `nl`, `ru`, `it`.
- `var_key` - (Optional) Used by node type: `PLAIN`.
- `var_value` - (Optional) Used by node type: `PLAIN`.
- `codes` - (Optional) Used by node type: `PIN_CODE`. A list of code objects with attributes `code` and `label`.
- `support_hours_id` - (Optional) Used by node type: `SUPPORT_HOURS`.
- `hold_audio_url` - (Optional) Used by node type: `ROUTE_CALL`.
- `targets` - (Optional) Used by node type: `ROUTE_CALL`. A list of targets with attributes `target` and `type`. `type` allowed values: `USER`, `ON_CALL_SCHEDULE`, `NUMBER`.
- `call_style` - (Optional) Used by node type: `ROUTE_CALL`. Allowed values: `ORDERED`, `RANDOM`, `PARALLEL`.
- `alert_source_id` - (Optional) Used by node type: `CREATE_ALERT`.
- `retries` - (Optional) Used by node types: `IVR_MENU`, `PIN_CODE`, `ROUTE_CALL`.
- `call_timeout_sec` - (Optional) Used by node type: `ROUTE_CALL`.
- `blacklist` - (Optional) Used by node type: `BLOCK_NUMBERS`.
- `intents` - (Optional) Used by node type: `AGENTIC`. A list of intent objects with attributes `type`, `label`, `description`, `examples`. `type` allowed values: `INCIDENT`, `SYSTEM_OUTAGE`, `SECURITY_BREACH`, `TECHNICAL_SUPPORT`, `INQUIRY`.
- `gathers` - (Optional) Used by node type: `AGENTIC`. A list of gather objects with attributes `type`, `label`, `var_type`, `required`, `question`. `type` allowed values: `CALLER_NAME`, `CONTACT_NUMBER`, `EMAIL`, `INCIDENT`, `AFFECTED_SERVICES`. `var_type` allowed values: `NUMBER`, `DATE`, `BOOLEAN`, `STRING`.
- `enrichment` - (Optional) Used by node type: `AGENTIC`. A single [enrichment](#enrichment-arguments) block.

##### Enrichment Arguments

- `enabled` - (Required)
- `information_types` - (Optional) A map of information types. Allowed values: `INCIDENT`, `MAINTENANCE`, `SERVICE_STATUS`.
- `sources` - (Optional) A map of sources with `id` and `type`. `type` allowed values: `STATUS_PAGE`, `SERVICE`.

#### Branch Arguments

- `branch_type` - (Required) The type of the branch. Allowed values: `BRANCH`, `CATCH_ALL`, `ANSWERED`.
- `condition` - (Optional) The branch condition.
- `target` - (Required) A single [node](#node-arguments) block.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the call flow.
- `assigned_number` - The assigned number object with `id`, `name` and nested `phone_number` containing `region_code` and `number`.

## Import

Call flows can be imported using the `id`, e.g.

```sh
$ terraform import ilert_call_flow.example 123456789
```
