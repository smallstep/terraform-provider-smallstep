terraform {
  required_providers {
    smallstep = {
      source = "smallstep/smallstep"
    }
  }
}

provider "smallstep" {
  client_certificate = {
    certificate = file("api.crt")
    private_key = file("api.key")
    team_id = "94a7dd82-1360-4493-b1bf-b14a97c45786"
  }
}

data "smallstep_authority" "preexisting" {
  id = "129dd44e-0381-4d45-a446-216b9a6e3bf4"
}

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

output "bootstrap_preexisting" {
  value = "step ca bootstrap --ca-url https://${data.smallstep_authority.preexisting.domain} --fingerprint ${data.smallstep_authority.preexisting.fingerprint} --context preexisting"
}
