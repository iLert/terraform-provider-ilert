# Alert Source Example

This demos [alert sources](https://docs.ilert.com/getting-started/intro#alert-source-inbound-integration).

This example will create an alert source in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_ORGANIZATION=
export ILERT_USERNAME=
export ILERT_PASSWORD=
```

```sh
terraform apply \
  -var "organization=${ILERT_ORGANIZATION}" \
  -var "username=${ILERT_USERNAME}" \
  -var "password=${ILERT_PASSWORD}" \
  -var "escalation_policy=123456789
```
