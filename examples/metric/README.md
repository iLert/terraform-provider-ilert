# Metric example

This demos [metrics](https://docs.ilert.com/incident-comms-and-status-pages/metrics).

This example will create a metric and a metric data source with its dependencies in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
