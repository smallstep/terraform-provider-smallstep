resource "smallstep_authority" "basic" {
  type              = "devops"
  active_revocation = false
  admin_emails      = ["eng@smallstep.com"]
  name              = "Basic"
  subdomain         = "basic"
}

output "bootstrap_basic" {
  value = "step ca bootstrap --ca-url https://${smallstep_authority.basic.domain} --fingerprint ${smallstep_authority.basic.fingerprint} --context basic"
}

resource "smallstep_authority" "advanced" {
  subdomain         = "myadvanced"
  name              = "My Advanced Authority"
  type              = "advanced"
  admin_emails      = ["eng@smallstep.com"]
  active_revocation = true
  intermediate_issuer = {
    name            = "My Custom Intermediate"
    key_version     = "RSA_SIGN_PKCS1_2048_SHA256"
    duration        = "100h"
    max_path_length = 0
    name_constraints = {
      critical                  = true
      permitted_ip_ranges       = ["10.32.0.0/12"]
      permitted_dns_domains     = [".cluster.local"]
      permitted_email_addresses = ["eng@smallstep.com"]
      permitted_uri_domains     = ["uri.cluster.local"]
    }
    subject = {
      common_name         = "Issuer"
      country             = "US"
      email_address       = "test@smallstep.com"
      locality            = "San Francisco"
      organization        = "Engineering"
      organizational_unit = "Core"
      postal_code         = "94108"
      province            = "CA"
      serial_number       = "1"
      street_address      = "26 O'Farrell St"
    }
  }
  root_issuer = {
    name            = "My Custom Root"
    key_version     = "RSA_SIGN_PKCS1_2048_SHA256"
    duration        = "1000h"
    max_path_length = "1"
    name_constraints = {
      critical                 = false
      excluded_ip_ranges       = ["10.96.0.0/12"]
      excluded_dns_domains     = ["example.com"]
      excluded_email_addresses = ["eng@example.com"]
      excluded_uri_domains     = ["uri:example.com"]
    }
  }
}
