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
