resource "ilert_support_hour" "example" {
  name     = "example"
  timezone = "Europe/Berlin"
  support_days {
    monday {
      start = "08:00"
      end   = "17:00"
    }

    tuesday {
      start = "08:00"
      end   = "17:00"
    }

    wednesday {
      start = "08:00"
      end   = "17:00"
    }

    thursday {
      start = "08:00"
      end   = "17:00"
    }

    friday {
      start = "08:00"
      end   = "17:00"
    }
  }
}

resource "ilert_support_hour" "example_windows" {
  name     = "example_windows"
  timezone = "Europe/Berlin"

  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "09:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "12:00"
    }
  }

  support_windows {
    from {
      day_of_week = "MONDAY"
      time        = "13:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "17:00"
    }
  }

  support_windows {
    from {
      day_of_week = "FRIDAY"
      time        = "17:00"
    }
    to {
      day_of_week = "MONDAY"
      time        = "08:00"
    }
  }
}
