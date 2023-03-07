resource "ilert_schedule" "this" {
  name     = var.name
  timezone = "Europe/Berlin"
  type     = "STATIC"
  shift {
    user  = ilert_user.this.id
    start = "2024-03-07T15:00:00Z"
    end   = "2024-03-08T15:00:00Z"
  }
  default_shift_duration = "PT24H"
}
