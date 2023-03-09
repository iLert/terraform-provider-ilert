resource "ilert_user" "example" {
  first_name = "example"
  last_name  = "example"
  email      = "example@example.com"
}

# example for recurring schedule
resource "ilert_schedule" "example_recurring" {
  name     = "example_recurring"
  timezone = "Europe/Berlin"
  type     = "RECURRING"
  schedule_layer {
    name      = "layer1"
    starts_on = formatdate("YYYY-MM-DD'T'hh:mm:ss", timeadd(timestamp(), "730h"))
    user {
      id = ilert_user.example.id
    }
    rotation         = "P1D"
    restriction_type = "TIMES_OF_WEEK"
    restriction {
      from {
        day_of_week = "MONDAY"
        time        = "13:00"
      }
      to {
        day_of_week = "MONDAY"
        time        = "16:00"
      }
    }
  }
}

# example for static schedule
resource "ilert_schedule" "example_static" {
  name     = "example_static"
  timezone = "Europe/Berlin"
  type     = "STATIC"
  shift {
    user  = ilert_user.example.id
    start = formatdate("YYYY-MM-DD'T'hh:mm:ss", timeadd(timestamp(), "730h"))
    end   = formatdate("YYYY-MM-DD'T'hh:mm:ss", timeadd(timestamp(), "754h"))
  }
}
