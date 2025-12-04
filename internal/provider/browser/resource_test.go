package browser

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccBrowserResource(t *testing.T) {
	cred1 := utils.NewCredential(t)
	cred2 := utils.NewCredential(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	name2 := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	minConfig := fmt.Sprintf(`
resource "smallstep_browser" "test" {
	name = %q
	match_addresses = ["https://example.com"]
}
`, name)

	fullConfig := fmt.Sprintf(`
resource "smallstep_browser" "test" {
	name = %q
	match_addresses = ["https://example.com", "https://*.example.com"]
	credentials = [%q, %q]
}
`, name2, *cred1.Id, *cred2.Id)

	emptyConfig := fmt.Sprintf(`
resource "smallstep_browser" "test" {
	name = %q
	match_addresses = ["https://*.example.com"]
	credentials = []
}`, name)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_browser.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_browser.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_browser.test", "match_addresses.#", "1"),
					helper.TestCheckResourceAttr("smallstep_browser.test", "match_addresses.0", "https://example.com"),
					helper.TestCheckResourceAttr("smallstep_browser.test", "credentials.#", "0"),
				),
			},
			{
				ResourceName:      "smallstep_browser.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fullConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_browser.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_browser.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_browser.test", "match_addresses.#", "2"),
					helper.TestCheckResourceAttr("smallstep_browser.test", "credentials.#", "2"),
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
					helper.TestMatchResourceAttr("smallstep_browser.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_browser.test", "name", name2),
					helper.TestCheckResourceAttr("smallstep_browser.test", "match_addresses.#", "2"),
					helper.TestCheckResourceAttr("smallstep_browser.test", "match_addresses.0", "https://example.com"),
					helper.TestCheckResourceAttr("smallstep_browser.test", "match_addresses.1", "https://*.example.com"),
					helper.TestCheckResourceAttr("smallstep_browser.test", "credentials.#", "2"),
					helper.TestMatchResourceAttr("smallstep_browser.test", "credentials.0", utils.UUIDRegexp),
					helper.TestMatchResourceAttr("smallstep_browser.test", "credentials.1", utils.UUIDRegexp),
				),
			},
			{
				ResourceName:      "smallstep_browser.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_browser.test", plancheck.ResourceActionUpdate),
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
					helper.TestMatchResourceAttr("smallstep_browser.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_browser.test", "name", name),
					helper.TestCheckResourceAttr("smallstep_browser.test", "match_addresses.#", "1"),
					helper.TestCheckResourceAttr("smallstep_browser.test", "credentials.#", "0"),
				),
			},
			{
				ResourceName:      "smallstep_browser.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: minConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_browser.test", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: emptyConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_browser.test", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
