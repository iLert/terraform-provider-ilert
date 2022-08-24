resource "ilert_user" "example" {
  email      = "example@example.com"
  username   = "example"
  first_name = "example"
  last_name  = "example"

  mobile {
    region_code = "DE"
    number      = "+491758250853"
  }

  high_priority_notification_preference {
    method = "EMAIL"
    delay  = 0
  }

  low_priority_notification_preference {
    method = "EMAIL"
    delay  = 0
  }

  on_call_notification_preference {
    method     = "EMAIL"
    before_min = 60
  }
}

