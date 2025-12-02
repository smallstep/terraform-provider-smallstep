
resource "smallstep_credential" "test" {
  slug = "slug"

  certificate = {
    authority_id = smallstep_authory.staging.id
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
