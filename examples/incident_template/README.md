# Incident template example

This demos [incident templates](https://docs.ilert.com/incident-comms-and-status-pages/incidents#incident-templates).

This example will create an incident template in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
