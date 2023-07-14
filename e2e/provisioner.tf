
resource "smallstep_provisioner" "my_oidc" {
  authority_id = smallstep_authority.advanced.id
  name         = "My OIDC example.com"
  type         = "OIDC"

  oidc = {
    client_id              = "abc123"
    client_secret          = "xyz789"
    configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
  }
}

resource "smallstep_provisioner" "jwk1" {
  authority_id = smallstep_authority.advanced.id
  name         = "jwk1"
  type         = "JWK"

  jwk = {
    key           = file("${path.module}/foo.pub")
    encrypted_key = file("${path.module}/foo.priv")
  }
}

resource "smallstep_provisioner" "my_attest" {
  authority_id = smallstep_authority.advanced.id
  name         = "attest foo"
  type         = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["apple", "step"]
  }
}
resource "smallstep_provisioner" "my_acme" {
  authority_id = smallstep_authority.advanced.id
  name         = "acme1"
  type         = "ACME"
  acme = {
    challenges  = ["http-01", "dns-01", "tls-alpn-01"]
    require_eab = true
    force_cn    = false
  }
}
