# if you want do use structure with groups, please do as following:
# 1. create service, create status page without structure (id reference to status page group is not yet available)
# 2. terraform apply
# 3. create data source to get created status page, create status page group, add structure block to the status page with group reference
# 4. terraform apply

resource "ilert_service" "example" {
  name = "example"
}

# data "ilert_status_page" "example" {
#   name = "example"
# }

# resource "ilert_status_page_group" "example" {
#   name = "example"
#   status_page {
#     id = data.ilert_status_page.example.id
#   }
# }

resource "ilert_status_page" "example" {
  name       = "example"
  subdomain  = "example.ilert.io"
  visibility = "PUBLIC"

  service {
    id = ilert_service.example.id
  }

  # structure {
  #   element {
  #     id   = ilert_status_page_group.example.id
  #     type = "GROUP"
  #     options = ["expand"]
  #     child {
  #       id   = ilert_service.example.id
  #       type = "SERVICE"
  #       options = ["no-graph"]
  #     }
  #   }
  #   element {
  #     id   = ilert_service.example.id
  #     type = "SERVICE"
  #   }
  # }
}
