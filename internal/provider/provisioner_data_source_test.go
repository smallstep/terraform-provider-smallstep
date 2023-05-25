package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_provisioner.test", "authority_id", authority.Id),
					resource.TestCheckResourceAttr("data.smallstep_provisioner.test", "name", provisioner.Name),
					resource.TestCheckResourceAttr("data.smallstep_provisioner.test", "id", *provisioner.Id),
					resource.TestCheckResourceAttr("data.smallstep_provisioner.test", "type", string(provisioner.Type)),
					resource.TestCheckResourceAttr("data.smallstep_provisioner.test", "oidc.client_id", oidc.ClientID),
					resource.TestCheckResourceAttr("data.smallstep_provisioner.test", "oidc.client_secret", oidc.ClientSecret),
					resource.TestCheckResourceAttr("data.smallstep_provisioner.test", "oidc.configuration_endpoint", oidc.ConfigurationEndpoint),
					resource.TestMatchResourceAttr("data.smallstep_provisioner.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
