---
layout: "ilert"
page_title: "ilert: ilert_metric_data_source"
sidebar_current: "docs-ilert-resource-metric-data-source"
description: |-
  Creates and manages a metric data source in ilert.
---

# ilert_metric_data_source

A [metric data source](https://api.ilert.com/api-docs/#tag/Metric-Data-Sources) is a resource for storing metadata (e.g. credentials) for a data provider. It is used for metrics to import data from third party tools.

## Example Usage

```hcl
resource "ilert_metric_data_source" "example_prometheus" {
  name = "example"
  type = "PROMETHEUS"

  metadata {
    auth_type  = "BASIC"
    basic_user = "your prometheus user name"
    basic_pass = "your prometheus password"
    url        = "your prometheus url"
  }
}

resource "ilert_metric_data_source" "example_datadog" {
  name = "example"
  type = "DATADOG"

  metadata {
    region          = "EU1"
    api_key         = "your datadog api key"
    application_key = "your datadog application key"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the metric data source.
- `type` - (Optional) The provider type of the metric data source. Allowed values are `DATADOG`, `PROMETHEUS`.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `metadata` - (Optional) A [metadata](#metadata-arguments) block.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### Metadata Arguments

- `region` - (Optional) The region of the provider. (Datadog)
- `api_key` - (Optional) The api key of the provider. (Datadog)
- `application_key` - (Optional) The application key of the provider. (Datadog)
- `auth_type` - (Optional) The auth type for the provider. Allowed values are `NONE`, `BASIC`, `HEADER`. (Prometheus)
- `basic_user` - (Optional) The username for the provider, required if `auth_type` is `BASIC`. (Prometheus)
- `basic_pass` - (Optional) The password for the provider, required if `auth_type` is `BASIC`. (Prometheus)
- `header_key` - (Optional) The custom key for the provider, required if `auth_type` is `HEADER`. (Prometheus)
- `header_value` - (Optional) The custom value for the provider, required if `auth_type` is `HEADER`. (Prometheus)
- `url` - (Optional) The url for the provider. (Prometheus)

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the metric data source.
- `name` - The name of the metric data source.

## Import

Metric data sources can be imported using the `id`, e.g.

```sh
$ terraform import ilert_metric_data_source.main 123456789
```
