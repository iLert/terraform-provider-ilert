resource "ilert_alert_action" "this" {
  name = "My GitHub Alert Action"

  alert_source {
    id = ilert_alert_source.this.id
  }

  connector {
    id   = ilert_connector.this.id
    type = ilert_connector.this.type
  }

  github {
    owner      = "fake"
    repository = "fake"
  }
}
