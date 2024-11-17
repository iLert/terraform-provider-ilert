---
layout: "ilert"
page_title: "ilert: ilert_deployment_pipeline"
sidebar_current: "docs-ilert-data-source-deployment-pipeline"
description: |-
  Get information about a deployment pipeline that you have created.
---

# ilert_deployment_pipeline

Use this data source to get information about a specific [deployment pipeline][1].

## Example Usage

```hcl
data "ilert_deployment_pipeline" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The deployment pipeline name to use to find a deployment pipeline in the ilert API.

## Attributes Reference

- `id` - The ID of the found deployment pipeline.
- `name` - The name of the found deployment pipeline.

[1]: https://api.ilert.com/api-docs/#tag/deployment-pipelines
