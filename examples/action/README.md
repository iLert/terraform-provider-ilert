# Action example

This demos actions and credentials, which are early access resources of the ilert API. They are disabled by default and have to be enabled in the provider configuration by the paths of their APIs under `/api/beta`, with `early_access_enabled = ["actions", "credentials"]`, which [`providers.tf`](./providers.tf) sets. ilert also enables actions account by account.

This example will create a credential holding a Discord webhook URL and an action that posts the high priority alerts of an alert source to it, built from the `discord_alert` template. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

The templates of an account are listed by the API at `GET /api/beta/actions/templates`, and what a template needs at `GET /api/beta/actions/templates/{key}`.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
export DISCORD_WEBHOOK_URL=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
  -var "alert_source_id=123456" \
  -var "discord_webhook_url=${DISCORD_WEBHOOK_URL}"
```
