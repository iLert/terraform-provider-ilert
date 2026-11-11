# Agent example

This demos agents and credentials, which are early access resources of the ilert API. They are disabled by default and have to be enabled with `early_access_enabled` in the provider configuration, which [`providers.tf`](./providers.tf) sets.

This example will create a credential holding a GitHub token and an agent in the specified organization, whose GitHub MCP server authenticates with that credential. See https://registry.terraform.io/providers/iLert/ilert/latest/docs for details on configuring [`providers.tf`](./providers.tf) accordingly.

The names of the MCP servers and their tools are listed by the API at `GET /api/beta/agentic-workforce/metadata/servers-and-tools`.

Alternatively, you may use variables passed via command line:

```sh
export ILERT_API_TOKEN=
export GITHUB_TOKEN=
```

```sh
terraform apply \
  -var "api_token=${ILERT_API_TOKEN}" \
  -var "github_token=${GITHUB_TOKEN}"
```
