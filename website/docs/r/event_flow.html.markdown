---
layout: "ilert"
page_title: "ilert: ilert_event_flow"
sidebar_current: "docs-ilert-resource-event-flow"
description: |-
  Creates and manages an event flow in ilert.
---

# ilert_event_flow

An event flow enables defining complex node-based workflows for event processing and routing.

## Example Usage

```hcl
resource "ilert_event_flow" "example" {
  name = "example"

  root_node {
    node_type = "ROOT"
    branches {
      branch_type = "ACCEPTED"
      target {
        name      = "Set variable"
        node_type = "PLAIN"
        metadata {
          var_key   = "context.event.summary"
          var_value = "event from terraform"
        }
      }
    }
  }
}
```

> Note: More examples for event flows can be found in main.tf under examples/event_flow within the provider.

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the event flow.
- `team` - (Optional) One or more [team](#team-arguments) blocks. The order in which the blocks are declared is not significant.
- `root_node` - (Required) A single [node](#node-arguments) block defining the root of the event flow.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### Node Arguments

- `name` - (Optional) The name of the node.
- `node_type` - (Required) The node type. Allowed values: `ROOT`, `PLAIN`, `SUPPORT_HOURS`, `ROUTE_EVENT`, `DEFINE_BRANCHES`, `WAIT`, `TRANSFORM`.
- `metadata` - (Optional) A single [metadata](#metadata-arguments) block.
- `branches` - (Optional) A list of [branch](#branch-arguments) blocks.

#### Metadata Arguments

- `var_key` - (Optional) Used by node type: `PLAIN`.
- `var_value` - (Optional) Used by node type: `PLAIN`.
- `support_hours_id` - (Optional) Used by node type: `SUPPORT_HOURS`.
- `alert_source_id` - (Optional) Used by node type: `ROUTE_EVENT`.
- `overwrite_priority` - (Optional) Used by node type: `ROUTE_EVENT`. Allowed values: `HIGH`, `LOW`.
- `escalation_policy_id` - (Optional) Used by node type: `ROUTE_EVENT`.
- `definitions` - (Optional) Used by node type: `DEFINE_BRANCHES`. A list of definition objects with attributes `branch_name` and `conditions`.
- `wait_for_duration` - (Optional) Used by node type: `WAIT` (for example `PT5M`).
- `wait_start_support_hours_id` - (Optional) Used by node type: `WAIT`.
- `wait_end_support_hours_id` - (Optional) Used by node type: `WAIT`.
- `condition` - (Optional) Used by node type: `TRANSFORM`.
- `rules` - (Optional) Used by node type: `TRANSFORM`. A list of rule objects with attributes:
  - `name` (Required)
  - `target` (Required)
  - `operator` (Required) allowed values: `SET`, `COPY`, `MAP`, `TEMPLATE`, `MERGE`, `APPEND_ARRAY`
  - `value` (Optional, string)
  - `source` (Optional)
  - `mapping` (Optional, map of string values)
  - `default` (Optional, string)
  - `properties` (Optional, map of string values)
  - `items` (Optional, list of maps of string values)

#### Branch Arguments

- `branch_type` - (Required) The type of the branch. Allowed values: `BRANCH`, `CATCH_ALL`, `ACCEPTED`.
- `condition` - (Optional) The branch condition.
- `target` - (Optional) A single [node](#node-arguments) block.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the event flow.

## Import

Event flows can be imported using the `id`, e.g.

```sh
$ terraform import ilert_event_flow.example 123456789
```
