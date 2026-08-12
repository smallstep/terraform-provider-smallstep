
resource "smallstep_workload" "nginx" {
  name          = "Nginx TLS"
  workload_type = "nginx"

  credentials = [{
    credential_id = smallstep_credential.server.id

    probes = [{
      target      = "app.local:443"
      protocol    = "TLS"
      server_name = "app.local"
    }]
  }]

  hooks = {
    sign = {
      after = ["systemctl reload nginx"]
      shell = "/bin/bash"
    }
    renew = {
      after = ["systemctl reload nginx"]
      shell = "/bin/bash"
    }
  }

  reload_info = {
    method    = "DBUS"
    unit_name = "nginx.service"
  }
}
