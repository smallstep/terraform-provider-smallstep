package provisioner

import (
	"fmt"
	"regexp"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccProvisionerDataSource(t *testing.T) {
	t.Parallel()
	authority := utils.NewAuthority(t)
	provisioner, oidc := utils.NewOIDCProvisioner(t, authority.Id)
	config := fmt.Sprintf(`
data "smallstep_provisioner" "test" {
	authority_id = %q
	name = %q
}`, authority.Id, provisioner.Name)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_provisioner.test", "authority_id", authority.Id),
					helper.TestCheckResourceAttr("data.smallstep_provisioner.test", "name", provisioner.Name),
					helper.TestCheckResourceAttr("data.smallstep_provisioner.test", "id", *provisioner.Id),
					helper.TestCheckResourceAttr("data.smallstep_provisioner.test", "type", string(provisioner.Type)),
					helper.TestCheckResourceAttr("data.smallstep_provisioner.test", "oidc.client_id", oidc.ClientID),
					helper.TestCheckResourceAttr("data.smallstep_provisioner.test", "oidc.client_secret", oidc.ClientSecret),
					helper.TestCheckResourceAttr("data.smallstep_provisioner.test", "oidc.configuration_endpoint", oidc.ConfigurationEndpoint),
					helper.TestMatchResourceAttr("data.smallstep_provisioner.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
