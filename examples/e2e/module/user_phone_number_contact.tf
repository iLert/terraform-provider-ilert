resource "ilert_user_phone_number_contact" "this" {
  region_code = "DE"
  target      = "+4915123456789"
  user {
    id = ilert_user.this.id
  }
}
