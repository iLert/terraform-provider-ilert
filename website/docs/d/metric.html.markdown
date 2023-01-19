---
layout: "ilert"
page_title: "ilert: ilert_metric"
sidebar_current: "docs-ilert-data-source-metric"
description: |-
  Get information about a metric that you have created.
---

# ilert_metric

Use this data source to get information about a specific [metric][1].

## Example Usage

```hcl
data "ilert_metric" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The metric name to use to find a metric in the ilert API.

## Attributes Reference

- `id` - The ID of the found metric.
- `name` - The name of the found metric.
- `aggregation_type` - The aggregation type of the found metric.
- `display_type` - The display type of the found metric.

[1]: https://api.ilert.com/api-docs/#tag/Metrics
