resource "ilert_schedule" "this" {
  name     = var.name
  timezone = "Europe/Berlin"
  type     = "STATIC"
  shift {
    user  = ilert_user.this.id
    start = formatdate("YYYY-MM-DD'T'hh:mm:ss", timeadd(timestamp(), "730h"))
    end   = formatdate("YYYY-MM-DD'T'hh:mm:ss", timeadd(timestamp(), "754h"))
  }
  default_shift_duration = "PT24H"
}
