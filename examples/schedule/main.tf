resource "ilert_user" "example" {
  username   = "example1"
  first_name = "exam"
  last_name  = "ple"
  email      = "example1@example.com"
}

resource "ilert_schedule" "example" {
  name     = "example"
  timezone = "Europe/Berlin"
  type     = "RECURRING"
  schedule_layer {
    name      = "layer1"
    starts_on = "2022-08-30T00:00"
    user {
      id = ilert_user.example.id
    }
    rotation = "P1D"
  }
}
