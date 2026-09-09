resource "ilert_telemetry_source" "example" {
  name        = "example"
  type        = "OTEL"
  description = "example ilert telemetry source"

  service_name_prefix = "prod-"

  labels = {
    env = "production"
  }
}

output "integration_key" {
  description = "the credential the OpenTelemetry collector authenticates with"
  value       = ilert_telemetry_source.example.integration_key
  sensitive   = true
}
