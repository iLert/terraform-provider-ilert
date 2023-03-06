resource "ilert_user" "example" {
  email      = "example2@example.com"
  first_name = "example"
  last_name  = "example"
}

resource "ilert_user_phone_number_contact" "example" {
  region_code = "DE"
  target      = "+4915123456789" // for best practice, use FQTN E.164 format
  user {
    id = ilert_user.example.id
  }
}
