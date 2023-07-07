
resource "smallstep_agent_configuration" "agent1" {
  name             = "Agent1"
  authority_id     = smallstep_authority.agents_authority.id
  provisioner_name = smallstep_provisioner.acme_attest.name
  attestation_slug = smallstep_attestation_authority.aa.slug
  depends_on       = [smallstep_provisioner.acme_attest]
}
