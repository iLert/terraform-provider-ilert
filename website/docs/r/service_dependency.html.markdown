---
layout: "ilert"
page_title: "ilert: ilert_service_dependency"
sidebar_current: "docs-ilert-resource-service-dependency"
description: |-
  Creates and manages a dependency between two services in ilert.
---

# ilert_service_dependency

A service dependency is one edge of the ilert service dependency graph: it records that one [service](https://docs.ilert.com/developer-docs/rest-api/api-reference/services) depends on another.

Dependencies are their own resource rather than a block on `ilert_service` because the API gives each one its own id and its own endpoints — the service create and update payloads do not accept dependencies at all — and because the only bulk write it offers replaces every dependency of a service at once, which would remove edges created outside the configuration.

## Example Usage

```hcl
resource "ilert_service" "api" {
  name = "api"
}

resource "ilert_service" "database" {
  name = "database"
}

resource "ilert_service_dependency" "api_on_database" {
  service_id        = ilert_service.api.id
  target_service_id = ilert_service.database.id
  notes             = "the api reads from the database"
  type              = "HARD"
}
```

## Argument Reference

The following arguments are supported:

- `service_id` - (Required) The ID of the service that depends on the target service. Changing this forces a new resource to be created.
- `target_service_id` - (Required) The ID of the service that is being depended upon. Changing this forces a new resource to be created.
- `type` - (Optional) The strength of the dependency. Allowed values are `HARD` and `SOFT`. Defaults to `HARD`, which is what the API applies when the field is omitted. Changing this forces a new resource to be created.
- `notes` - (Optional) Notes describing the dependency. Changing this forces a new resource to be created.

Every argument forces a new resource because the API has no update for a single edge. Its bulk endpoint can change a type or a note in place, but it rewrites every dependency of the service at once and would drop edges this resource does not know about. Recreating one edge deletes and re-adds it, which changes nothing about either service.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the dependency, in the form `<service id>/<dependency id>`. The second half is the ID of the edge itself, not the ID of either service.
- `invalid_after` - Point in time after which the dependency is no longer considered valid, as a date time string in ISO format. Read-only: the API documents it as writable but ignores it on both create and bulk replace, in every date format, so it cannot be set from Terraform.
- `created_at` - The creation date of the dependency in ISO format.
- `updated_at` - The date the dependency was last updated, in ISO format.

## Import

Service dependencies can be imported using the `<service id>/<dependency id>` pair, e.g.

```sh
$ terraform import ilert_service_dependency.main 123456789/987654321
```
