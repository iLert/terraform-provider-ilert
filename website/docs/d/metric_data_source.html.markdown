---
layout: "ilert"
page_title: "ilert: ilert_metric_data_source"
sidebar_current: "docs-ilert-data-source-metric-data-source"
description: |-
  Get information about a metric data source that you have created.
---

# ilert_metric_data_source

Use this data source to get information about a specific [metric data source][1].

## Example Usage

```hcl
data "ilert_metric_data_source" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The metric data source name to use to find a metric data source in the ilert API.

## Attributes Reference

- `id` - The ID of the found metric data source.
- `name` - The name of the found metric data source.
- `type` - The provider type of the found metric data source.

[1]: https://api.ilert.com/api-docs/#tag/Metric-Data-Sources
