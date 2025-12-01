
resource "smallstep_managed_radius" "my_radius" {
  name      = "My RADIUS"
  nas_ips   = ["1.2.3.4"]
  client_ca = file("${path.module}/root.crt")
  reply_attributes = [{
    name  = "Tunnel-Type"
    value = "13"
    }, {
    name                   = "Tunnel-Private-Group-ID"
    value_from_certificate = "2.5.4.11"
  }]
}
