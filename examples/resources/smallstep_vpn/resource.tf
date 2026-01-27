
resource "smallstep_vpn" "my_vpn" {
  name            = "Employee VPN"
  autojoin        = true
  remote_address  = "10.20.30.40"
  connection_type = "IKEv2"
  ike = {
    ca_chain  = smallstep_authority.vpn.root
    eap       = true
    remote_id = "vpn.example.com"
  }
  credentials = [smallstep_credential.device.id]
}
