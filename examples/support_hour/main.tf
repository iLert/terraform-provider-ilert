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
