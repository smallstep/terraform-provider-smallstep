
resource "smallstep_provisioner" "my_oidc" {
  authority_id = smallstep_authority.basic.id
  name         = "My OIDC example.com"
  type         = "OIDC"

  oidc = {
    client_id              = "abc123"
    client_secret          = "xyz789"
    configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
  }
}

resource "smallstep_provisioner" "my_jwk" {
  authority_id = smallstep_authority.basic.id
  name         = "jwk1"
  type         = "JWK"

  jwk = {
    key           = file("${path.module}/foo.pub")
    encrypted_key = file("${path.module}/foo.priv")
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

resource "smallstep_provisioner" "my_x5c" {
  authority_id = smallstep_authority.basic.id
  name         = "x5c foo"
  type         = "X5C"
  x5c = {
    roots = ["-----BEGIN CERTIFICATE-----\n..."]
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
  name         = "attest foo"
  type         = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["tpm"]
    attestation_roots   = ["-----BEGIN CERTIFICATE-----\n..."]
  }
}
