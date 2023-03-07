# Alert action & connector example

This demos [alert actions and connectors](https://docs.ilert.com/getting-started/readme#connectors-and-alert-actions-aka-outbound-integrations).

This example will create an alert source, a connector, an alert action and its dependencies in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
