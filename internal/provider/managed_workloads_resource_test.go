package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/stretchr/testify/require"
)

func TestAccAgentConfigurationResource(t *testing.T) {
	attestorRoot, _ := utils.CACerts(t)
	slug := utils.Slug(t)
	config := fmt.Sprintf(`
resource "smallstep_collection" "tpms" {
	slug = %q
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
	attestor_roots = %q
	depends_on = [smallstep_collection.tpms]
}

resource "smallstep_authority" "agents" {
	subdomain = %q
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
`, slug, attestorRoot, slug)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_agent_configuration.agent1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
				),
			},
		},
	})
}
