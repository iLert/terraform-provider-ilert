resource "ilert_connector" "this" {
  name = var.name
  type = "github"

  github {
    api_key = "example"
  }
}
