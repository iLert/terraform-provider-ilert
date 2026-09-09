resource "ilert_user" "example" {
  email              = "example@example.com"
  first_name         = "example"
  last_name          = "example"
  send_no_invitation = true

  # uncomment to buy a license for this user instead of checking the account quota,
  # this always buys a seat and charges the account for it
  # purchase_seat = true
}
