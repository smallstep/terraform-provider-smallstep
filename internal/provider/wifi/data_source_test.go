package wifi

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccWifiDataSource(t *testing.T) {
	cred1 := utils.NewCredential(t)
	cred2 := utils.NewCredential(t)
	root, _ := utils.CACerts(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	ssid := "ssid-" + acctest.RandStringFromCharSet(8, acctest.CharSetAlphaNum)

	config := fmt.Sprintf(`
resource "smallstep_wifi" "test" {
	name = %q
	ssid = %q
	radius_server_ca = %q
	radius_server_domain = "radius.example.com"
	hidden = true
	autojoin = true
	credentials = [%q, %q]
}

data "smallstep_wifi" "test" {
	id = smallstep_wifi.test.id
}
`, name, ssid, root, *cred1.Id, *cred2.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("data.smallstep_wifi.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("data.smallstep_wifi.test", "name", name),
					helper.TestCheckResourceAttr("data.smallstep_wifi.test", "ssid", ssid),
					helper.TestCheckResourceAttr("data.smallstep_wifi.test", "radius_server_ca", root),
					helper.TestCheckResourceAttr("data.smallstep_wifi.test", "radius_server_domain", "radius.example.com"),
					helper.TestCheckResourceAttr("data.smallstep_wifi.test", "hidden", "true"),
					helper.TestCheckResourceAttr("data.smallstep_wifi.test", "autojoin", "true"),
					helper.TestCheckResourceAttr("data.smallstep_wifi.test", "credentials.#", "2"),
					helper.TestMatchResourceAttr("data.smallstep_wifi.test", "credentials.0", utils.UUIDRegexp),
					helper.TestMatchResourceAttr("data.smallstep_wifi.test", "credentials.1", utils.UUIDRegexp),
				),
			},
		},
	})
}
