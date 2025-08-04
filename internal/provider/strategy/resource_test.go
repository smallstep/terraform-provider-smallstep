package strategy

import (
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const browserConfig = `
	resource "smallstep_strategy" "browser" {
		name = "Browser Certificate"
		browser = {}
	}
`

func TestAccStrategyBrowser(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: browserConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_strategy.browser", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_strategy.browser", "name", "Browser Certificate"),
					helper.TestMatchResourceAttr("smallstep_strategy.browser", "certificate.authority_id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_strategy.browser", "browser.%", "0"),
				),
			},
		},
	})
}
