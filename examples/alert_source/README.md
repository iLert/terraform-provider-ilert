# Alert source example

This demos [alert sources](https://docs.ilert.com/getting-started/readme#alert-source-aka-inbound-integration).

This example will create three different alert sources and its dependencies in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```

## Migrating legacy email alert source to the new experience

Previously an alert source of type `EMAIL` was configured via a variety of fields, including:

- `email_filtered`
- `email_resolve_filtered`
- `filter_operator`
- `resolve_filter_operator`
- `resolve_key_extractor`
- `email_predicate`
- `email_resolve_predicate`

Those fields are now deprecated and replaced with templated fields in the new email alert source with type `EMAIL2`:

- `alert_key_template`
- `event_type_filter_create`
- `event_type_filter_accept`
- `event_type_filter_resolve`

They are configured using the [ilert ITL](https://docs.ilert.com/rest-api/icl-ilert-condition-language) and work well with, but are not limited to alert sources of type `EMAIL2`.

### Example

Old configuration:

```hcl
resource "ilert_alert_source" "example_email" {
  name              = "My Email Integration from terraform"
  integration_type  = "EMAIL"
  email             = "example@ 'your tenant' .ilert.eu"
  escalation_policy = ilert_escalation_policy.example.id

  alert_creation = "OPEN_RESOLVE_ON_EXTRACTION"
  resolve_key_extractor {
    field    = "EMAIL_SUBJECT"
    criteria = "ALL_TEXT_BEFORE"
    value    = "my server"
  }

  email_filtered = true
  email_predicate {
    field    = "EMAIL_BODY"
    criteria = "CONTAINS_STRING"
    value    = "alarm"
  }

  email_resolve_filtered = true
  email_resolve_predicate {
    field    = "EMAIL_BODY"
    criteria = "CONTAINS_STRING"
    value    = "resolve"
  }
}
```

New configuration:

```hcl
resource "ilert_alert_source" "example_email" {
  name              = "My Email Integration from terraform"
  integration_type  = "EMAIL2"
  email             = "example@ 'your tenant' .ilert.eu"
  escalation_policy = ilert_escalation_policy.example.id

  alert_creation = "OPEN_RESOLVE_ON_EXTRACTION"
  alert_key_template {
    text_template = "{{ subject.splitTakeAt(\"my server\", 0) }}"
  }

  event_type_filter_create  = "(event.customDetails.body contains_any [\"alarm\"])"
  event_type_filter_resolve = "(event.customDetails.body in [\"resolve\"])"
}
```
