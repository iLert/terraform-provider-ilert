# Service dependency example

This demos the edges of the [service dependency graph](https://docs.ilert.com/developer-docs/rest-api/api-reference/services).

This example will create two services in the specified organization and record that one depends on the other. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
