
resource "smallstep_provisioner_webhook" "external" {
	authority_id = smallstep_authority.foo.id
	provisioner_id = smallstep_provisioner.bar.id
	name = "devices"
	kind = "ENRICHING"
	cert_type = "X509"
	server_type = "EXTERNAL"
	url = "https://example.com/hook"
	bearer_token = "secret123"
}

resource "smallstep_provisioner_webhook" "devices" {
	authority_id = smallstep_authority.foo.id
	provisioner_id = smallstep_provisioner.bar.id
	name = "devices"
	kind = "ENRICHING"
	cert_type = "X509"
	server_type = "HOSTED_ATTESTATION"
	collection_slug = smallstep_collection.tpms.slug
	depends_on = [smallstep_collection.tpms]
}

