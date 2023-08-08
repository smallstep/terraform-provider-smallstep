package agent_configuration

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccAgentConfigurationDataSource(t *testing.T) {
	authority := utils.NewAuthority(t)
	provisioner, _ := utils.NewOIDCProvisioner(t, authority.Id)
	ac := utils.NewAgentConfiguration(t, authority.Id, provisioner.Name, "attestslug")

	agentConfig := fmt.Sprintf(`
data "smallstep_agent_configuration" "agent" {
	id = %q
}
`, *ac.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: agentConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "id", *ac.Id),
					helper.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "authority_id", authority.Id),
					helper.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "name", ac.Name),
					helper.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "provisioner_name", ac.Provisioner),
					helper.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "attestation_slug", *ac.AttestationSlug),
				),
			},
		},
	})
}
