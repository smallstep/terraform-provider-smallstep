resource "smallstep_credential" "basic" {
  slug = "workstation-cred"

  certificate = {
    authority_id = smallstep_authority.staging.id
    duration     = "168h"
    x509 = {
      common_name = {
        device_metadata = "smallstep:identity"
      }
      sans = {
        device_metadata = ["smallstep:identity"]
      }
    }
  }

  key = {
    type       = "ECDSA_P384"
    protection = "HARDWARE_ATTESTED"
  }

  policy = {
    os        = ["Linux"]
    ownership = ["company"]
  }

  files = {
    root_file = "/var/ssl/ca.pem"
  }
}

resource "smallstep_credential" "advanced" {
  slug            = "server-cred"
  management_mode = "agent"

  certificate = {
    authority_id = smallstep_authority.staging.id
    duration     = "24h"

    x509 = {
      common_name = {
        static = "server.example.com"
      }
      organization = {
        static = ["Example Inc"]
      }
      organizational_unit = {
        device_metadata            = ["department"]
        insecure_include_requested = true
      }
      typed_sans = {
        dns_names = {
          static          = ["server.example.com"]
          device_metadata = ["Device.Hostname"]
        }
        ip_addresses = {
          device_metadata = ["Device.PermanentIdentifier"]
        }
      }
      extended_key_usage = ["serverAuth", "clientAuth"]
      custom_extensions = [{
        oid   = "1.3.6.1.4.1.44947"
        value = "dGVzdA=="
      }]
    }

    name_policy = {
      allow = {
        dns = ["*.internal.example.com"]
      }
      allow_wildcard_names = true
    }
  }

  key = {
    type          = "ECDSA_P384"
    protection    = "HARDWARE_ATTESTED"
    compatibility = "DEFAULT"
    store         = "MACHINE"
  }

  policy = {
    assurance = ["high"]
    os        = ["Linux"]
    ownership = ["company"]
  }

  files = {
    root_file  = "/var/ssl/ca.pem"
    crt_file   = "/var/ssl/cert.pem"
    key_file   = "/var/ssl/key.pem"
    key_format = "DEFAULT"
    mode       = 384 # 0600
  }
}
