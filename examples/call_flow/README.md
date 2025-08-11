# Call flow example

This demos [call flows](https://docs.ilert.com/call-routing/call-routing-2.0-beta).

This example will create a call flow with a root node and two branches to showcase the structure. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
```
