# User phone number contact example

This demos [user phone number contacts](https://docs.ilert.com/getting-started/readme#notifications).

This example will create a user with a user phone number contact in the specified organization. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}"
```
