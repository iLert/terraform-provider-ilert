---
layout: "ilert"
page_title: "iLert: ilert_incident_template"
sidebar_current: "docs-ilert-data-source-incident-template"
description: |-
  Get information about an incident template that you have created.
---

# ilert_incident_template

Use this data source to get information about a specific [incident template][1].

## Example Usage

```hcl
data "ilert_incident_template" "example" {
  name = "example"
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The incident template name to use to find an incident template in the iLert API.

## Attributes Reference

- `id` - The ID of the found incident template.
- `name` - The name of the found incident template.
- `status` - The status of the found incident template.
- `summary` - The summary of the found incident template.
- `message` - The message of the found incident template.

[1]: https://api.ilert.com/api-docs/#tag/Incident-Templates
