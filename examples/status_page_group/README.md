# Status Page Group Example

This demos [status page groups](https://docs.ilert.com/incident-comms-and-status-pages/status-pages).

This example will create a service, a status page group and a status page containing the group including the service in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}"
```
