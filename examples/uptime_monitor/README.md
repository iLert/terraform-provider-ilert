# Uptime monitor example

This demos uptime monitors.

> Uptime monitors are soon to be deprecated and should only be used when they are really needed.

This example will create a user, an escalation policy and an uptime monitor in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}"
```
