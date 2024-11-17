---
layout: "ilert"
page_title: "ilert: ilert_deployment_pipeline"
sidebar_current: "docs-ilert-resource-deployment-pipeline"
description: |-
  Creates and manages a deployment pipeline in ilert.
---

# ilert_deployment_pipeline

A [deployment pipeline](https://api.ilert.com/api-docs/#tag/deployment-pipelines) enables you to send deployment events to ilert.

## Example Usage

```hcl
resource "ilert_deployment_pipeline" "example" {
  name     = "example"
  integration_type = "GITHUB"
  github {
    branch_filter = ["main", "master"]
    event_filter = ["release"]
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the deployment pipeline.
- `integration_type` - (Required) The integration type of the deployment pipeline. Allowed values are `GITHUB` or `API`.
- `integration_key` - (Computed) The integration key of the deployment pipeline.
- `team` - (Optional) One or more [team](#team-arguments) blocks.
- `created_at` - (Computed) The creation date of the deployment pipeline.
- `updated_at` - (Computed) The latest date the deployment pipeline was updated at.
- `integration_url` - (Computed) The integration url of the deployment pipeline. Deployment events are sent to this URL.
- `github` - The [github](#github-arguments) block allows configuring parameters for GitHub.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

#### GitHub Arguments

- `branch_filter` - One or more branch filters to only accept events on specified branches.
- `event_filter` - One or more event filters to only accept events on specified actions. Allowed values are `pull_request`, `push` and `release`.

## Import

Deployment pipelines can be imported using the `id`, e.g.

```sh
$ terraform import ilert_deployment_pipeline.main 123456789
```
