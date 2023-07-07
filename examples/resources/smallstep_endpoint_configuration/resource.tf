
resource "smallstep_endpoint_configuration" "x509" {
  name = "My DB"
  kind = "WORKLOAD"

  authority_id     = smallstep_authority.endpoints.id
  provisioner_name = smallstep_provisioner.endpoints_x5c.name

  certificate_info = {
    type      = "X509"
    duration  = "168h"
    crt_file  = "db.crt"
    key_file  = "db.key"
    root_file = "ca.crt"
    uid       = 1001
    gid       = 999
    mode      = 256
  }

  hooks = {
    renew = {
      shell    = "/bin/sh"
      before   = ["echo renewing"]
      after    = ["echo renewed"]
      on_error = ["echo failed renew"]
    }
    sign = {
      shell    = "/bin/bash"
      before   = ["echo signing"]
      after    = ["echo signed"]
      on_error = ["echo failed sign"]
    }
  }

  key_info = {
    format   = "DEFAULT"
    type     = "ECDSA_P256"
    pub_file = "file.csr"
  }

  reload_info = {
    method   = "SIGNAL"
    pid_file = "db.pid"
    signal   = 1
  }
}

resource "smallstep_endpoint_configuration" "ssh" {
  name             = "SSH"
  kind             = "PEOPLE"
  authority_id     = smallstep_authority.endpoints.id
  provisioner_name = smallstep_provisioner.endpoints_x5c.name
  certificate_info = {
    type = "SSH_USER"
  }
  key_info = {
    type   = "RSA_2048"
    format = "OPENSSH"
  }
}
