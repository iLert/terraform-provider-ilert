resource "ilert_team" "example" {
  name = "example"
}

resource "ilert_metric_data_source" "example_prometheus" {
  name = "example"
  type = "PROMETHEUS"

  team {
    id   = ilert_team.example.id
    name = "example"
  }

  metadata {
    auth_type  = "BASIC"
    basic_user = "your prometheus user name"
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
