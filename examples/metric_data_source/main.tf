resource "ilert_metric_data_source" "example_prometheus" {
  name = "example"
  type = "PROMETHEUS"

  metadata {
    auth_type  = "BASIC"
    basic_user = "your prometheus user name"
    basic_pass = "your prometheus password"
    url        = "your prometheus url"
  }
}

resource "ilert_metric_data_source" "example_datadog" {
  name = "example"
  type = "DATADOG"

  metadata {
    region          = "EU1"
    api_key         = "your datadog api key"
    application_key = "your datadog application key"
  }
}
