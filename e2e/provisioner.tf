data "smallstep_provisioner" "prov" {
  authority_id = "129dd44e-0381-4d45-a446-216b9a6e3bf4"
  id = "af7f3b22-057c-4bd8-8832-bca9f0c59aa8"
}

output "provisioner_name" {
  value = data.smallstep_provisioner.prov.name
}

resource "smallstep_provisioner" "my_oidc" {
  authority_id = smallstep_authority.basic.id
  name = "My OIDC example.com"
  type = "OIDC"

  oidc = {
    client_id = "abc123"
    client_secret = "xyz789"
    configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
  }
}

resource "smallstep_provisioner" "jwk1" {
  authority_id = smallstep_authority.basic.id
  name = "jwk1"
  type = "JWK"

  jwk = {
    key = file("${path.module}/foo.pub")
    encrypted_key = file("${path.module}/foo.priv")
  }
}

resource "smallstep_provisioner" "my_x5c" {
  authority_id = smallstep_authority.basic.id
  name         = "x5c foo"
  type         = "X5C"
  x5c = {
    roots = ["-----BEGIN CERTIFICATE-----\nMIIBazCCARGgAwIBAgIQIbPd9RVuu0l/eTKVVjDL9zAKBggqhkjOPQQDAjAUMRIw\nEAYDVQQDEwlyb290LWNhLTEwHhcNMjIxMjE0MjIzMzE0WhcNMzIxMjExMjIzMzE0\nWjAUMRIwEAYDVQQDEwlyb290LWNhLTEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\nAASIOTDFWdDWOIyzASpf9EKEokDkc1ckVt50pjgkalI9wQ8WyH88xutdUNDjIVDK\nVryJaH2DMUdWwFh07kmEyu8co0UwQzAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/\nBAgwBgEB/wIBATAdBgNVHQ4EFgQUq5LjtW5K1PrPYV8mJjcCTvuKvyYwCgYIKoZI\nzj0EAwIDSAAwRQIhAJRTP0IUbOVmW2ufj8VINXnBhYCCauSa1/fHNNM/5o1JAiAx\n+Pkrk5eAijv8H5xDEW2ik3bUDnai7rQLzdbrEb5RPw==\n-----END CERTIFICATE-----"]
  }
}

resource "smallstep_provisioner" "my_attest" {
  authority_id = smallstep_authority.basic.id
  name         = "attest foo"
  type         = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["apple", "step"]
  }
}

resource "smallstep_provisioner" "tpm_attest" {
  authority_id = smallstep_authority.basic.id
  name         = "attest tpm"
  type         = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["tpm"]
    attestation_roots   = ["-----BEGIN CERTIFICATE-----\nMIIBazCCARGgAwIBAgIQIbPd9RVuu0l/eTKVVjDL9zAKBggqhkjOPQQDAjAUMRIw\nEAYDVQQDEwlyb290LWNhLTEwHhcNMjIxMjE0MjIzMzE0WhcNMzIxMjExMjIzMzE0\nWjAUMRIwEAYDVQQDEwlyb290LWNhLTEwWTATBgcqhkjOPQIBBggqhkjOPQMBBwNC\nAASIOTDFWdDWOIyzASpf9EKEokDkc1ckVt50pjgkalI9wQ8WyH88xutdUNDjIVDK\nVryJaH2DMUdWwFh07kmEyu8co0UwQzAOBgNVHQ8BAf8EBAMCAQYwEgYDVR0TAQH/\nBAgwBgEB/wIBATAdBgNVHQ4EFgQUq5LjtW5K1PrPYV8mJjcCTvuKvyYwCgYIKoZI\nzj0EAwIDSAAwRQIhAJRTP0IUbOVmW2ufj8VINXnBhYCCauSa1/fHNNM/5o1JAiAx\n+Pkrk5eAijv8H5xDEW2ik3bUDnai7rQLzdbrEb5RPw==\n-----END CERTIFICATE-----"]
  }
}
resource "smallstep_provisioner" "my_acme" {
  authority_id = smallstep_authority.basic.id
  name         = "acme1"
  type         = "ACME"
  acme = {
    challenges  = ["http-01", "dns-01", "tls-alpn-01"]
    require_eab = true
    force_cn    = false
  }
}
