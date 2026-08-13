
resource "smallstep_proxy" "corp_proxy" {
  name            = "CorpProxy"
  remote_address  = "proxy.example.com:3128"
  credentials     = [smallstep_credential.device.id]
  match_addresses = ["*.internal.example.com", "10.0.0.0/8"]
}
