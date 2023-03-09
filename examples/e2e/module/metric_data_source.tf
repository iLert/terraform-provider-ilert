resource "ilert_metric_data_source" "this" {
  name = var.name
  type = "DATADOG"

  team {
    id = ilert_team.this.id
  }

  metadata {
    region          = "EU1"
    api_key         = "example"
    application_key = "example"
  }
}
