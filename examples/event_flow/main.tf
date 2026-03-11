resource "ilert_event_flow" "example" {
  name = "example"

  root_node {
    node_type = "ROOT"
    branches {
      branch_type = "ACCEPTED"
      target {
        name      = "Transform event"
        node_type = "TRANSFORM"
        metadata {
          condition = "context.event != null"
          rules {
            name     = "Rule 1"
            target   = "context.event.summary"
            operator = "SET"
            value    = "managed by terraform"
          }
        }
      }
    }
  }
}
