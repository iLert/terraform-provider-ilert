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

If you want to create a **status page with groups**, please follow the following instructions:

- 1. create service, create status page without structure (id reference to status page group is not yet available)
- 2. terraform apply
- 3. create data source to get created status page, create status page group, add structure block to the status page with group reference
- 4. terraform apply

> If you have already created a status page or a status page with a status page group, the steps above are not needed.

When destroying the resources it is recommended to destroy status page groups first and status page second to ensure a correct Terraform state.
