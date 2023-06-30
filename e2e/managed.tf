
resource "smallstep_collection" "tpms" {
	slug = "tpms"
}

resource "smallstep_collection_instance" "server1" {
	id = "urn:ek:sha256:RAzbOveN1Y45fYubuTxu5jOXWtOK1HbfZ7yHjBuWlyE="
	data = "{}"
	collection_slug = smallstep_collection.tpms.slug
	depends_on = [smallstep_collection.tpms]
}

resource "smallstep_attestation_authority" "aa" {
	name = "foo"
	catalog = smallstep_collection.tpms.slug
  attestor_roots = file("${path.module}/ca.crt")
	depends_on = [smallstep_collection.tpms]
}

resource "smallstep_authority" "agents" {
	subdomain = "agents"
	name = "Agents Authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "agents" {
	authority_id = smallstep_authority.agents.id
	name = "Agents"
	type = "ACME_ATTESTATION"
	acme_attestation = {
		attestation_formats = ["tpm"]
		attestation_roots = [smallstep_attestation_authority.aa.root]
	}
}

resource "smallstep_provisioner_webhook" "devices" {
	authority_id = smallstep_authority.agents.id
	provisioner_id = smallstep_provisioner.agents.id
	name = "devices"
	kind = "ENRICHING"
	cert_type = "X509"
	server_type = "HOSTED_ATTESTATION"
	collection_slug = smallstep_collection.tpms.slug
	depends_on = [smallstep_collection.tpms]
}

resource "smallstep_agent_configuration" "agent1" {
	authority_id = smallstep_authority.agents.id
	provisioner_name = smallstep_provisioner.agents.name
	name = "Agent1"
	attestation_slug = smallstep_attestation_authority.aa.slug
	depends_on = [smallstep_provisioner.agents]
}
