data "ilert_metric_data_source" "example" {
  name = "example"
}

resource "ilert_metric" "example" {
  name             = "example"
  aggregation_type = "AVG"
  display_type     = "GRAPH"
  metadata {
    query = "your query"
  }
  data_source {
    id = data.ilert_metric_data_source.example.id
  }
}
