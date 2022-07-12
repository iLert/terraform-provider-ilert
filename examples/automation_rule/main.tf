data "ilert_service" "example" {
  name = "example"
}

data "ilert_alert_source" "example" {
  name = "example"
}

resource "ilert_automation_rule" "example" {
  name = "example"
  alert_type = "CREATED"
  service_status = "OPERATIONAL"
  service = data.ilert_service.example.id
  alert_source = data.ilert_alert_source.example.id
}
