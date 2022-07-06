# Connection Example

> Legacy API - please use alert-actions - for more information see https://docs.ilert.com/rest-api/api-version-history#renaming-connections-to-alert-actions

This demos [connections and connectors](https://docs.ilert.com/getting-started/intro#connectors-and-connections-outbond-integrations).

This example will create an alert source, a connector and connection in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

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
  -var "password=${ILERT_PASSWORD}"
```
