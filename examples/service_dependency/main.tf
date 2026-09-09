resource "ilert_service" "api" {
  name        = "api"
  description = "the api depends on the database below"
}

resource "ilert_service" "database" {
  name = "database"
}

resource "ilert_service_dependency" "api_on_database" {
  service_id        = ilert_service.api.id
  target_service_id = ilert_service.database.id
  notes             = "the api reads from the database"
  type              = "HARD"
}
