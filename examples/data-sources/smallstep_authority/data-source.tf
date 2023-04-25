data "smallstep_authority" "preexisting" {
  id = "34bd2a7f-68e5-4f5e-81a5-531a4c3b5d99"
}

output "bootstrap_preexisting" {
  value = "step ca bootstrap --ca-url https://${data.smallstep_authority.preexisting.domain} --fingerprint ${data.smallstep_authority.preexisting.fingerprint} --context preexisting"
}
