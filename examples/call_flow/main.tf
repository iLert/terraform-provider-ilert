resource "ilert_call_flow" "example" {
  name     = "example"
  language = "en"

  root_node {
    node_type = "ROOT"
    branches {
      branch_type = "ANSWERED"
      target {
        name = "Create alert"
        node_type = "CREATE_ALERT"
        metadata {
          alert_source_id = -1 // your alert source id
        }
      }
    }
  }
}
