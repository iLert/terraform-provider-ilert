resource "ilert_user" "example" {
  username   = "example1"
  first_name = "John"
  last_name  = "Doe"
  email      = "john.doe@example.com"
}

# example for recurring schedule
resource "ilert_schedule" "example_recurring" {
  name     = "example_recurring"
  timezone = "Europe/Berlin"
  type     = "RECURRING"
  schedule_layer {
    name      = "layer1"
    starts_on = "2022-08-30T00:00"
    user {
      id = ilert_user.example.id
    }
    rotation         = "P1D"
    restriction_type = "TIME_OF_WEEK"
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
  team {
    id = 0
  }
}

# example for static schedule
resource "ilert_schedule" "example_static" {
  name     = "example_static"
  timezone = "Europe/Berlin"
  type     = "STATIC"
  shift {
    user  = ilert_user.example.id
    start = "2022-09-01T08:00"
    end   = "2022-09-02T08:00"
  }
  team {
    id = 0
  }
  default_shift_duration = "PT24H"
}
