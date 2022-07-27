resource "ilert_incident_template" "example" {
  name              = "example"
  status            = "INVESTIGATING"
  send_notification = true
}
