resource "ilert_user" "this" {
  email      = "${var.name}@fake.com"
  username   = var.name
  first_name = "fake"
  last_name  = "fake"

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

