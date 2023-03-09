# Escalation policy example

This demos [escalation policies](https://docs.ilert.com/getting-started/readme#escalation-policy).

This example will create three different alert sources and its dependencies in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
