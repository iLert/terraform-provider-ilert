resource "ilert_credential" "github" {
  name         = "github"
  type         = "BEARER"
  secret_value = var.github_token
  category     = "MCP"
  sub_category = "github"
}

resource "ilert_agent" "example" {
  name = "example"

  mcp_server {
    name          = "ilert-workforce-github-mcp"
    credential_id = ilert_credential.github.id
  }

  mcp_tool {
    server = "ilert-workforce-github-mcp"
    name   = "get_pull_request"

    input_overwrite {
      name   = "owner"
      values = ["iLert"]
    }

    input_overwrite {
      name   = "repo"
      values = ["ilert-go", "terraform-provider-ilert"]
    }
  }
}
