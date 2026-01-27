package vpn

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccVpnDataSource(t *testing.T) {
	root, _ := utils.CACerts(t)
	name := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	remoteAddress := "10.0.0.1"

	config := fmt.Sprintf(`
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
}

data "smallstep_vpn" "test" {
	id = smallstep_vpn.test.id
}
`, name, remoteAddress, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("data.smallstep_vpn.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("data.smallstep_vpn.test", "name", name),
					helper.TestCheckResourceAttr("data.smallstep_vpn.test", "connection_type", "IKEv2"),
					helper.TestCheckResourceAttr("data.smallstep_vpn.test", "remote_address", remoteAddress),
					helper.TestCheckResourceAttr("data.smallstep_vpn.test", "autojoin", "true"),
					helper.TestCheckResourceAttr("data.smallstep_vpn.test", "ike.ca_chain", root),
					helper.TestCheckResourceAttr("data.smallstep_vpn.test", "ike.eap", "true"),
					helper.TestCheckResourceAttr("data.smallstep_vpn.test", "ike.remote_id", "vpn.example.com"),
				),
			},
		},
	})
}
