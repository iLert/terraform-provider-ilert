resource "ilert_deployment_pipeline" "example" {
  name     = "example"
  integration_type = "GITHUB"
  github {
    branch_filter = ["main", "master"]
    event_filter = ["release"]
  }
}
