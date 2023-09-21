package agent_configuration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/authority"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/provisioner"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
		authority.NewResource,
		provisioner.NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccAgentConfigurationResource(t *testing.T) {
	root, _ := utils.CACerts(t)
	slug := utils.Slug(t)
	config1 := fmt.Sprintf(`
resource "smallstep_authority" "agents" {
	subdomain = %q
	name = "tfprovider-agents-authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "agents" {
	authority_id = smallstep_authority.agents.id
	name = "Agents"
	type = "X5C"
	x5c = {
		roots = [%q]
	}
}

resource "smallstep_agent_configuration" "agent1" {
	authority_id = smallstep_authority.agents.id
	provisioner_name = smallstep_provisioner.agents.name
	name = "tfprovider Agent1"
	attestation_slug = "anythinggoes"
	depends_on = [smallstep_provisioner.agents]
}`, slug, root)

	// update authority, provisioner, slug without recreate
	slug2 := utils.Slug(t)
	config2 := fmt.Sprintf(`
resource "smallstep_authority" "agents" {
	subdomain = %q
	name = "tfprovider-agents-authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "agents" {
	authority_id = smallstep_authority.agents.id
	name = "Agents"
	type = "X5C"
	x5c = {
		roots = [%q]
	}
}

resource "smallstep_agent_configuration" "agent1" {
	authority_id = smallstep_authority.agents.id
	provisioner_name = smallstep_provisioner.agents.name
	name = "tfprovider Agent 1"
	attestation_slug = "anythinggoes2"
	depends_on = [smallstep_provisioner.agents]
}`, slug2, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config1,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_agent_configuration.agent1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_agent_configuration.agent1", "name", "tfprovider Agent1"),
					helper.TestMatchResourceAttr("smallstep_agent_configuration.agent1", "authority_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_agent_configuration.agent1", "provisioner_name", "Agents"),
					helper.TestCheckResourceAttr("smallstep_agent_configuration.agent1", "attestation_slug", "anythinggoes"),
				),
			},
			{
				Config: config2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_agent_configuration.agent1", "name", "tfprovider Agent 1"),
				),
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_agent_configuration.agent1", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      "smallstep_agent_configuration.agent1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
