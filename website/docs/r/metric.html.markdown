---
layout: "ilert"
page_title: "ilert: ilert_metric"
sidebar_current: "docs-ilert-resource-metric"
description: |-
  Creates and manages a metric in ilert.
---

# ilert_metric

A [metric](https://api.ilert.com/api-docs/#tag/Metric) lets you provide additional information about the health of your services in status pages.

## Example Usage

```hcl
resource "ilert_metric_data_source" "example" {
  name = "example"
  type = "PROMETHEUS"

  metadata {
    auth_type  = "BASIC"
    basic_user = "your prometheus username"
    basic_pass = "your prometheus password"
    url        = "your prometheus url"
  }
}

resource "ilert_metric" "example" {
  name             = "example"
  aggregation_type = "AVG"
  display_type     = "GRAPH"
  metadata {
    query = "your query"
  }
  data_source {
    id = ilert_metric_data_source.example.id
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the metric.
- `description` - (Optional) The description for the metric.
- `aggregation_type` - (Required) The aggregation type of the metric. Allowed values are `AVG`, `SUM`, `MIN`, `MAX`, `LAST`.
- `display_type` - (Required) The display type of the metric. Allowed values are `GRAPH`, `SINGLE`.
- `interpolate_gaps` - (Optional) Indicates whether or not gaps will be interpolated.
- `lock_y_axis_max` - (Optional) The maximum value at which the graph is locked.
- `lock_y_axis_min` - (Optional) The minimum value at which the graph is locked.
- `mouse_over_decimal` - (Optional) Determines how many decimals are shown when hovering over the graph.
- `show_values_on_mouse_over` - (Optional) Indicates whether or not values will be shown when hovering over the graph.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `unit_label` - (Optional) The unit label of the metric.
- `metadata` - (Optional, Required with `data_source`) A [metadata](#metadata-arguments) block.
- `data_source` - (Optional, Required with `metadata`) A [data source](#data-source-arguments) block.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### Metadata Arguments

- `query` - (Required) The query used for the metric. (Datadog, Prometheus)

#### Data source Arguments

- `id` - (Required) The id of the metric data source used for the metric.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the found metric.
- `name` - The name of the found metric.
- `aggregation_type` - The aggregation type of the found metric.
- `display_type` - The display type of the found metric.

## Import

Metrics can be imported using the `id`, e.g.

```sh
$ terraform import ilert_metric.main 123456789
```
