
resource "smallstep_wifi" "my_wireless_net" {
  name                 = "Employee Net"
  ssid                 = "employees"
  radius_server_ca     = smallstep_managed_radius.my_radius.server_ca
  radius_server_domain = smallstep_managed_radius.my_radius.server_hostname
  hidden               = false
  autojoin             = true
  credentials          = [smallstep_credential.device.id]
}
