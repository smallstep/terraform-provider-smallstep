package ethernet

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccEthernetDataSource(t *testing.T) {
	cred1 := utils.NewCredential(t)
	cred2 := utils.NewCredential(t)
	root, _ := utils.CACerts(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
resource "smallstep_ethernet" "test" {
	name = %q
	radius_server_ca = %q
	credentials = [%q, %q]
	autojoin = true
}

data "smallstep_ethernet" "test" {
	id = smallstep_ethernet.test.id
}
`, name, root, *cred1.Id, *cred2.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("data.smallstep_ethernet.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("data.smallstep_ethernet.test", "name", name),
					helper.TestCheckResourceAttr("data.smallstep_ethernet.test", "radius_server_ca", root),
					helper.TestCheckResourceAttr("data.smallstep_ethernet.test", "autojoin", "true"),
					helper.TestCheckResourceAttr("data.smallstep_ethernet.test", "credentials.#", "2"),
					helper.TestMatchResourceAttr("data.smallstep_ethernet.test", "credentials.0", utils.UUIDRegexp),
					helper.TestMatchResourceAttr("data.smallstep_ethernet.test", "credentials.1", utils.UUIDRegexp),
				),
			},
		},
	})
}
