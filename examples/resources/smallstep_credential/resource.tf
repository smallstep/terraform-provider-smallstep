
resource "smallstep_credential" "test" {
  slug = "slug"

  certificate = {
    authority_id = smallstep_authority.staging.id
    duration     = "168h"
    x509 = {
      common_name = {
        device_metadata = "smallstep:identity"
      }
      typed_sans = {
        dns_names = {
          device_metadata = ["Device.Hostname"]
        }
        user_principal_names = {
          device_metadata = ["smallstep:identity"]
        }
      }
      extended_key_usage = ["serverAuth", "clientAuth"]
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
