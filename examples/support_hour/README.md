# Support hours example

This demos [support hours](https://docs.ilert.com/alerting/support-hours).

This example will create two support hours resources in the specified organization, one with a window per day and one with a split shift and coverage across the weekend. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
