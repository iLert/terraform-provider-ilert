---
layout: "ilert"
page_title: "Provider: ilert"
sidebar_current: "docs-ilert-index"
description: |-
  Terraform provider ilert.
---

# ilert Provider

The ilert provider is used to interact with ilert resources.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
provider "ilert" {
    api_token = "your api token, excluding Bearer prefix"
}

# Example resource configuration
resource "ilert_alert_source" "example" {
  # ...
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

- `api_token` - (Optional) An ilert OAuth / Personal Access Token. When not provided or made available via the `ILERT_API_TOKEN` environment variable, the provider can only access resources available anonymously. Conflicts with `organization`. Make sure to exclude the `Bearer` prefix.

- `organization` - (Optional) This is the target ilert organization account to manage. It is optional to provide this value and it can also be sourced from the `ILERT_ORGANIZATION` environment variable. For example, `ilert` is a valid organization. Conflicts with `api_token` and requires `username` and `password`, as the individual account corresponding to provided `username` and `password` will need "owner" privileges for this organization.

- `username` - (Optional) An ilert username. When not provided or made available via the `ILERT_USERNAME` environment variable.

- `password` - (Optional) An ilert password. When not provided or made available via the `ILERT_PASSWORD` environment variable.

- `endpoint` - (Optional) This is the target ilert base API endpoint. Providing a value is a requirement when working with ilert Enterprise. It is optional to provide this value and it can also be sourced from the `ILERT_ENDPOINT` environment variable. The value must end with a slash, for example: `https://ilert.example.com/`
