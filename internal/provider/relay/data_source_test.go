package relay

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccRelayDataSource(t *testing.T) {
	relayHostname := utils.RelayHostname()
	relayName := "tfprovider-" + utils.Slug(t)
	caCert, _ := utils.CACerts(t)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(`
resource "smallstep_relay" "test" {
  name                  = %[1]q
  hostname              = %[3]q
  ca_chain              = %[2]q
  allowed_targets       = ["target1.example.com", "target2.example.com"]
}

data "smallstep_relay" "test" {
  id = smallstep_relay.test.id
}
`, relayName, caCert, relayHostname),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttrSet("data.smallstep_relay.test", "id"),
					helper.TestCheckResourceAttr("data.smallstep_relay.test", "name", relayName),
					helper.TestCheckResourceAttr("data.smallstep_relay.test", "hostname", relayHostname),
					helper.TestCheckResourceAttr("data.smallstep_relay.test", "ca_chain", caCert),
					helper.TestCheckResourceAttr("data.smallstep_relay.test", "allowed_targets.#", "2"),
					helper.TestCheckResourceAttr("data.smallstep_relay.test", "allowed_targets.0", "target1.example.com"),
					helper.TestCheckResourceAttr("data.smallstep_relay.test", "allowed_targets.1", "target2.example.com"),
				),
			},
		},
	})
}
