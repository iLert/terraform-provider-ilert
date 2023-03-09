# Schedule example

This demos [schedules](https://docs.ilert.com/on-call-management-and-escalations/on-call-schedules).

This example will create a static schedule and a recurring schedule with its dependencies in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
