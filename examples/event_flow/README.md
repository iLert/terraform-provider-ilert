# Event flow example

This demos [event flows](https://docs.ilert.com/event-workflows/).

This example will create an event flow with a root node and one branch to showcase the structure. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
