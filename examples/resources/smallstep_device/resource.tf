
resource "smallstep_device" "laptop1" {
  permanent_identifier = "782BF520"
  display_id           = "99-TUR"
  display_name         = "Employee Laptop"
  serial               = "6789"
  tags                 = ["ubuntu"]
  metadata = {
    k1 = "v1"
  }
  os        = "Linux"
  ownership = "company"
  user = {
    email = "user@example.com"
  }
}
