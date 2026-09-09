# Telemetry source example

This demos [telemetry sources](https://docs.ilert.com/developer-docs/rest-api/api-reference/telemetry-sources).

This example will create a telemetry source in the specified organization, which receives OpenTelemetry data and turns it into ilert services. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

The integration key the collector authenticates with is exported as a sensitive output, read it with `terraform output -raw integration_key`.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
