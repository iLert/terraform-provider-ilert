---
layout: "ilert"
page_title: "iLert: ilert_uptime_monitor"
sidebar_current: "docs-ilert-resource-uptime-monitor"
description: |-
  Creates and manages an uptime monitor in iLert.
---

# ilert_uptime_monitor

An [uptime monitor](https://api.ilert.com/api-docs/#tag/Uptime-Monitors) allows you to quickly setup monitoring for any kind of exposed service e.g. HTTP (e.g. websites), ICMP (ping) or TCP and UDP servers.

## Example Usage

```hcl
data "ilert_escalation_policy" "default" {
  name = "Default"
}

resource "ilert_uptime_monitor" "example" {
  name              = "example.com"
  region            = "EU"
  escalation_policy = data.ilert_escalation_policy.default.id
  interval_sec      = 900
  timeout_ms        = 10000
  check_type        = "http"

  check_params {
    url = "https://example.com"
  }
}

```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the uptime monitor.
- `region` - (Optional) The region of the uptime monitor. Allowed values are `EU` and `US`. Default: `EU`.
- `check_type` - (Required) The check type of the uptime monitor. Allowed values are `http`, `ping`, `tcp`, `ssl` and `udp`.
- `check_params` - (Required) A [check params](#check-params-arguments) block.
- `interval_sec` - (Optional) The check interval in seconds of the uptime monitor. Allowed values are `60`, `300`, `600`, `900`, `1800` and `3600`. Default: `300`.
- `timeout_ms` - (Optional) The check timeout in milliseconds of the uptime monitor. Allowed values are between `1000` to `60000`. Default: `30000`.
- `create_incident_after_failed_checks` - (Optional) The incident creation ratio after failed checks of the uptime monitor. Allowed values are between `1` to `12`. Default: `1`.
- `escalation_policy` - (Required) The escalation policy id used by this uptime monitor.
- `paused` - (Optional) The paused state of the uptime monitor. Default: `false`.

#### Check Params Arguments

- `host` - (Optional) The host name to check. This option is required if `check_type` is `ping`, `tcp` or `udp`.
- `port` - (Optional) The host port to check. This option is required if `check_type` is `tcp` or `udp`.
- `url` - (Optional) The url to check. This option is required if `check_type` is `http`.
- `response_keywords` - (Optional) The response keywords to check in the response body. This option is only used for `http`.
- `alert_before_sec` - (Optional) Time in seconds to alert before the certificate expires. This option is only used for `ssl`.
- `alert_on_fingerprint_change` - (Optional) Enables alerts when the certificate fingerprint changes. This option is only used for `ssl`.

## Attributes Reference

The following attributes are exported:

- `id` - The ID of the uptime monitor.
- `name` - The name of the uptime monitor.
- `status` - The status of the uptime monitor.
- `embed_url` - The embed report url of the uptime monitor.
- `shared_url` - The shared report url of the uptime monitor.

## Import

Services can be imported using the `id`, e.g.

```sh
$ terraform import ilert_uptime_monitor.main 123456789
```
