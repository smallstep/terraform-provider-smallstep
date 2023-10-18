
resource "smallstep_workload" "redis" {
  depends_on             = [smallstep_device_collection.ec2_west]
  device_collection_slug = resource.smallstep_device_collection.ec2_west.slug
  workload_type          = "redis"
  slug                   = "redisec2west"
  display_name           = "Redis EC2 West"
  admin_emails           = ["admin@example.com"]

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
      shell = "/bin/sh"
      before = [
        "echo renewing",
      ]
      after = [
        "echo renewed",
      ]
      on_error = [
        "echo failed renew",
      ]
    }
    sign = {
      shell = "/bin/bash"
      before = [
        "echo signing",
      ]
      after = [
        "echo signed",
      ]
      on_error = [
        "echo failed sign",
      ]
    }
  }

  key_info = {
    format = "DEFAULT"
    type   = "ECDSA_P256"
  }

  reload_info = {
    method   = "SIGNAL"
    pid_file = "db.pid"
    signal   = 1
  }
}
