resource "ilert_incident_template" "this" {
  name              = var.name
  status            = "INVESTIGATING"
  send_notification = true
  summary           = "example summary"
}
