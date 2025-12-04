package wifi

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccWifiResource(t *testing.T) {
	cred1 := utils.NewCredential(t)
	cred2 := utils.NewCredential(t)
	root, _ := utils.CACerts(t)
	root2, _ := utils.CACerts(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	name2 := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	ssid := "ssid-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)
	ssid2 := "ssid-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	minConfig := fmt.Sprintf(`
resource "smallstep_wifi" "test" {
	name = %q
	ssid = %q
	radius_server_ca = %q
}
`, name, ssid, root)

	fullConfig := fmt.Sprintf(`
resource "smallstep_wifi" "test" {
	name = %q
	ssid = %q
	radius_server_ca = %q
	radius_server_domain = "radius.example.com"
	hidden = true
	autojoin = true
	credentials = [%q, %q]
}
`, name2, ssid2, root2, *cred1.Id, *cred2.Id)

	emptyConfig := fmt.Sprintf(`
resource "smallstep_wifi" "test" {
	name = %q
	ssid = %q
	radius_server_ca = %q
	radius_server_domain = ""
	hidden = false
	autojoin = false
	credentials = []
}`, name, ssid, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_wifi.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "ssid", ssid),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "credentials.#", "0"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "radius_server_ca", root),
				),
			},
			{
				ResourceName:      "smallstep_wifi.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fullConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_wifi.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_wifi.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "ssid", ssid2),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "hidden", "true"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "credentials.#", "2"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "radius_server_domain", "radius.example.com"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "radius_server_ca", root2),
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
					helper.TestMatchResourceAttr("smallstep_wifi.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "ssid", ssid2),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "hidden", "true"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "credentials.#", "2"),
					helper.TestMatchResourceAttr("smallstep_wifi.test", "credentials.0", utils.UUIDRegexp),
					helper.TestMatchResourceAttr("smallstep_wifi.test", "credentials.1", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "radius_server_domain", "radius.example.com"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "radius_server_ca", root2),
				),
			},
			{
				ResourceName:      "smallstep_wifi.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_wifi.test", plancheck.ResourceActionUpdate),
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
					helper.TestMatchResourceAttr("smallstep_wifi.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "ssid", ssid),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "hidden", "false"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "autojoin", "false"),
					helper.TestCheckResourceAttr("smallstep_wifi.test", "credentials.#", "0"),
				),
			},
			{
				ResourceName:      "smallstep_wifi.test",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_wifi.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: emptyConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_wifi.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
