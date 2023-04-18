terraform {
  required_providers {
    smallstep = {
      source = "smallstep/smallstep"
    }
  }
}

provider "smallstep" {}

data "smallstep_authority" "authority" {
  id = "34bd2a7f-68e5-4f5e-81a5-531a4c3b5d99"
}

resource "smallstep_authority" "basic" {
  type = "devops"
  active_revocation = false
  admin_emails = ["andrew@smallstep.com"]
  name = "Basic"
  subdomain = "basic"
}

output "my_ca" {
  value = data.smallstep_authority.authority.domain
}

output "my_fingerprint" {
  value = data.smallstep_authority.authority.fingerprint
}
