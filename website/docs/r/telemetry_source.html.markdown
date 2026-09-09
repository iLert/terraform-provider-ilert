---
layout: "ilert"
page_title: "ilert: ilert_telemetry_source"
sidebar_current: "docs-ilert-resource-telemetry-source"
description: |-
  Creates and manages a telemetry source in ilert.
---

# ilert_telemetry_source

A [telemetry source](https://docs.ilert.com/developer-docs/rest-api/api-reference/telemetry-sources) receives OpenTelemetry data and turns it into ilert services, so the health of the instrumented system is reflected without a service having to be created by hand for each of its parts.

## Example Usage

```hcl
resource "ilert_telemetry_source" "example" {
  name = "example"
  type = "OTEL"
  description = "example ilert telemetry source"

  service_name_prefix = "prod-"

  labels = {
    env = "production"
  }
}

output "otel_integration_key" {
  value     = ilert_telemetry_source.example.integration_key
  sensitive = true
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the telemetry source.
- `type` - (Required) The telemetry protocol of the source. Allowed values are `OTEL`. Changing this forces a new resource to be created, the protocol is immutable after the source was created.
- `description` - (Optional) The description of the telemetry source.
- `service_name_prefix` - (Optional) Prepended to the name of every service ilert creates from the telemetry of this source, for example to keep the services of a staging environment apart from production. Leading and trailing whitespace is removed and no separator is inserted, so include one if you want one. Services that already exist are not renamed. Maximum 64 characters.
- `service_name_suffix` - (Optional) Appended to the name of every service ilert creates from the telemetry of this source, under the same rules as `service_name_prefix`. If prefix, service name and suffix together exceed the 255 character service name limit, the service name is shortened so that both affixes are kept in full.
- `labels` - (Optional) A map of free-form key-value labels assigned to the telemetry source. The labels are managed by Terraform only once the attribute is declared: removing it after it was declared clears the labels on the server, while a configuration that never declared it leaves labels assigned elsewhere untouched. The API stamps a reserved `ilert.com/discovered-by` label onto every telemetry source, derived from its name; it is read-only server-managed metadata, so the provider never sends it and keeps it out of the state. Declaring it in a configuration has no effect.
- `team` - (Optional) One or more [team](#team-arguments) blocks. The order in which the blocks are declared is not significant. Teams follow the same rule as `labels`: they are managed by Terraform only once a block is declared.

#### Team Arguments

- `id` - (Required) The ID of the team.
- `name` - (Optional) The name of the team.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the telemetry source.
- `name` - The name of the telemetry source.
- `status` - Whether ilert is currently receiving telemetry from this source. One of `PENDING`, `RECEIVING`, `STOPPED`, `ERROR`.
- `integration_key` - The credential the collector authenticates with when sending telemetry to ilert. Marked sensitive, so an `output` referencing it needs `sensitive = true`. The API omits it for users whose API key has no update permission on the telemetry source; in that case the value already in state is kept rather than emptied.
- `created_at` - The creation date of the telemetry source in ISO format.
- `updated_at` - The date the telemetry source was last updated, in ISO format.

## Import

Telemetry sources can be imported using the `id`, e.g.

```sh
$ terraform import ilert_telemetry_source.main 123456789
```

`labels` and `team` are not captured by the import, matching the way `team` behaves on every other resource: both are managed only once declared, and the import has no configuration to read that from. The first plan after importing re-asserts whatever the configuration declares.

Note that the API answers a read of a telemetry source your API token cannot see with `404` rather than `403`, which is indistinguishable from a deleted one. Refreshing with a token that lacks access to the source — one scoped to teams the source is not assigned to, for example — drops it from the state, and the next apply creates a second one. Use a token with access to every telemetry source the configuration manages.
