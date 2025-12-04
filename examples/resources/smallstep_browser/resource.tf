
resource "smallstep_browser" "intranet" {
  name            = "Intranet"
  match_addresses = ["https://smallstep.internal"]
  credentials     = [smallstep_credential.device.id]
}
