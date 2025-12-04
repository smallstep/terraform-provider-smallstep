package ethernet

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccEthernetResource(t *testing.T) {
	cred1 := utils.NewCredential(t)
	cred2 := utils.NewCredential(t)
	root, _ := utils.CACerts(t)
	root2, _ := utils.CACerts(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	name2 := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	minConfig := fmt.Sprintf(`
resource "smallstep_ethernet" "test" {
	name = %q
	radius_server_ca = %q
}
`, name, root)

	fullConfig := fmt.Sprintf(`
resource "smallstep_ethernet" "test" {
	name = %q
	radius_server_ca = %q
	credentials = [%q, %q]
	autojoin = true
}
`, name2, root2, *cred1.Id, *cred2.Id)

	emptyConfig := fmt.Sprintf(`
resource "smallstep_ethernet" "test" {
	name = %q
	radius_server_ca = %q
	credentials = []
	autojoin = false
}`, name, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_ethernet.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "credentials.#", "0"),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "radius_server_ca", root),
				),
			},
			{
				ResourceName:      "smallstep_ethernet.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fullConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_ethernet.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "credentials.#", "2"),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "radius_server_ca", root2),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "autojoin", "true"),
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
					helper.TestMatchResourceAttr("smallstep_ethernet.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "credentials.#", "2"),
					helper.TestMatchResourceAttr("smallstep_ethernet.test", "credentials.0", utils.UUIDRegexp),
					helper.TestMatchResourceAttr("smallstep_ethernet.test", "credentials.1", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "radius_server_ca", root2),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "autojoin", "true"),
				),
			},
			{
				ResourceName:      "smallstep_ethernet.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_ethernet.test", plancheck.ResourceActionUpdate),
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
					helper.TestMatchResourceAttr("smallstep_ethernet.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "credentials.#", "0"),
					helper.TestCheckResourceAttr("smallstep_ethernet.test", "autojoin", "false"),
				),
			},
			{
				ResourceName:      "smallstep_ethernet.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_ethernet.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: emptyConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_ethernet.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
