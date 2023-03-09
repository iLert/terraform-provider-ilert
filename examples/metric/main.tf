resource "ilert_metric_data_source" "example" {
  name = "example"
  type = "PROMETHEUS"

  metadata {
    auth_type  = "BASIC"
    basic_user = "your prometheus username"
    basic_pass = "your prometheus password"
    url        = "your prometheus url"
  }
}

resource "ilert_metric" "example" {
  name             = "example"
  aggregation_type = "AVG"
  display_type     = "GRAPH"
  metadata {
    query = "your query"
  }
  data_source {
    id = ilert_metric_data_source.example.id
  }
}
