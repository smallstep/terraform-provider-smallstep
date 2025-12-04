package browser

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccBrowserDataSource(t *testing.T) {
	cred1 := utils.NewCredential(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
resource "smallstep_browser" "test" {
	name = %q
	match_addresses = ["https://example.com", "https://*.example.com"]
	credentials = [%q]
}

data "smallstep_browser" "test" {
	id = smallstep_browser.test.id
}
`, name, *cred1.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("data.smallstep_browser.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("data.smallstep_browser.test", "name", name),
					helper.TestCheckResourceAttr("data.smallstep_browser.test", "match_addresses.#", "2"),
					helper.TestCheckResourceAttr("data.smallstep_browser.test", "match_addresses.0", "https://example.com"),
					helper.TestCheckResourceAttr("data.smallstep_browser.test", "match_addresses.1", "https://*.example.com"),
					helper.TestCheckResourceAttr("data.smallstep_browser.test", "credentials.#", "1"),
					helper.TestMatchResourceAttr("data.smallstep_browser.test", "credentials.0", utils.UUIDRegexp),
				),
			},
		},
	})
}
