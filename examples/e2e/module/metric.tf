resource "ilert_metric" "this" {
  name             = var.name
  aggregation_type = "AVG"
  display_type     = "GRAPH"
  metadata {
    query = "fake"
  }
  data_source {
    id = ilert_metric_data_source.this.id
  }
}
