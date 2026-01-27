
resource "smallstep_ethernet" "my_wired_net" {
  name             = "CorpNet"
  radius_server_ca = smallstep_managed_radius.my_radius.server_ca
  autojoin         = true
  credentials      = [smallstep_credential.device.id]
}
