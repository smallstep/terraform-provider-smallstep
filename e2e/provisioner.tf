
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

resource "smallstep_provisioner" "my_aws" {
  authority_id = smallstep_authority.advanced.id
  name         = "AWS foo"
  type         = "AWS"
  aws = {
    accounts                   = ["0123456789"]
    instance_age               = "1h"
    disable_trust_on_first_use = true
    disable_custom_sans        = true
  }
}
resource "smallstep_provisioner" "my_gcp" {
  authority_id = smallstep_authority.advanced.id
  name         = "GCP foo"
  type         = "GCP"
  gcp = {
    project_ids                = ["prod-1234"]
    service_accounts           = ["pki@prod-1234.iam.gserviceaccount.com"]
    instance_age               = "1h"
    disable_trust_on_first_use = true
    disable_custom_sans        = true
  }
}

resource "smallstep_provisioner" "my_azure" {
  authority_id = smallstep_authority.advanced.id
  name         = "Azure foo"
  type         = "AZURE"
  azure = {
    tenant_id                  = "948920a7-8fc1-431f-be94-030a80232e51"
    resource_groups            = ["prod-1234"]
    audience                   = "example.com"
    disable_trust_on_first_use = true
    disable_custom_sans        = true
  }
}
