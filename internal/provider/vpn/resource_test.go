package vpn

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccVpnResource(t *testing.T) {
	cred1 := utils.NewCredential(t)
	cred2 := utils.NewCredential(t)
	root, _ := utils.CACerts(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	name2 := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	remoteAddress := "10.0.0.1"
	remoteAddress2 := "10.0.0.2"

	minConfig := fmt.Sprintf(`
resource "smallstep_vpn" "test" {
	name = %q
	connection_type = "IPsec"
	remote_address = %q
}
`, name, remoteAddress)

	fullConfig := fmt.Sprintf(`
resource "smallstep_vpn" "test" {
	name = %q
	connection_type = "IKEv2"
	remote_address = %q
	autojoin = true
	ike = {
		ca_chain = %q
		eap = true
		remote_id = "vpn.example.com"
	}
	credentials = [%q, %q]
}
`, name2, remoteAddress2, root, *cred1.Id, *cred2.Id)

	emptyConfig := fmt.Sprintf(`
resource "smallstep_vpn" "test" {
	name = %q
	connection_type = "IKEv2"
	remote_address = %q
	autojoin = false
	ike = {
		ca_chain = %q
		eap = false
		remote_id = ""
	}
	credentials = []
}`, name, remoteAddress, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_vpn.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "connection_type", "IPsec"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "remote_address", remoteAddress),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "credentials.#", "0"),
				),
			},
			{
				ResourceName:      "smallstep_vpn.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fullConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_vpn.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_vpn.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "connection_type", "IKEv2"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "remote_address", remoteAddress2),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.ca_chain", root),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.eap", "true"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.remote_id", "vpn.example.com"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "credentials.#", "2"),
				),
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_vpn.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "connection_type", "IKEv2"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "remote_address", remoteAddress2),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "credentials.#", "2"),
					helper.TestMatchResourceAttr("smallstep_vpn.test", "credentials.0", utils.UUIDRegexp),
					helper.TestMatchResourceAttr("smallstep_vpn.test", "credentials.1", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.ca_chain", root),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.eap", "true"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.remote_id", "vpn.example.com"),
				),
			},
			{
				ResourceName:      "smallstep_vpn.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_vpn.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_vpn.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "connection_type", "IKEv2"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "remote_address", remoteAddress),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "autojoin", "false"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "credentials.#", "0"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.eap", "false"),
					helper.TestCheckResourceAttr("smallstep_vpn.test", "ike.remote_id", ""),
				),
			},
			{
				ResourceName:      "smallstep_vpn.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_vpn.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: emptyConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_vpn.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
