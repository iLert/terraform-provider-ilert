---
layout: "ilert"
page_title: "Provider: iLert"
sidebar_current: "docs-ilert-index"
description: |-
  Terraform provider iLert.
---

# iLert Provider

The iLert provider is used to interact with iLert resources.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "ilert" {
    organization = "your organization"
    username     = "your username"
    password     = "password"
}

# Example resource configuration
resource "alert_source_resource" "example" {
  # ...
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

- `api_token` - (Optional) A iLert OAuth / Personal Access Token. When not provided or made available via the `ILERT_API_TOKEN` environment variable, the provider can only access resources available anonymously. Conflicts with `organization`

- `organization` - (Optional) This is the target iLert organization account to manage. It is optional to provide this value and it can also be sourced from the `ILERT_ORGANIZATION` environment variable. For example, `ilert` is a valid organization. Conflicts with `api_token` and requires `username` and `password`, as the individual account corresponding to provided `username` and `password` will need "owner" privileges for this organization.

- `username` - (Optional) A iLert username. When not provided or made available via the `ILERT_USERNAME` environment variable.

- `password` - (Optional) A iLert password. When not provided or made available via the `ILERT_PASSWORD` environment variable.

- `endpoint` - (Optional) This is the target iLert base API endpoint. Providing a value is a requirement when working with iLert Enterprise. It is optional to provide this value and it can also be sourced from the `ILERT_ENDPOINT` environment variable. The value must end with a slash, for example: `https://ilert.example.com/`
