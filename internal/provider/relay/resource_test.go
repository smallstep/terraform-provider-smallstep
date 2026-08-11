package relay

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccRelayResource(t *testing.T) {
	relayHostname := utils.RelayHostname()
	relayHostname2 := utils.RelayHostname2()
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
  allowed_targets       = ["target1.example.com"]
}
`, relayName, caCert, relayHostname),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_relay.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_relay.test", "name", relayName),
					helper.TestCheckResourceAttr("smallstep_relay.test", "hostname", relayHostname),
					helper.TestCheckResourceAttr("smallstep_relay.test", "ca_chain", caCert),
					helper.TestCheckResourceAttr("smallstep_relay.test", "allowed_targets.0", "target1.example.com"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "smallstep_relay" "test" {
  name                  = %[1]q
  hostname              = %[3]q
  ca_chain              = %[2]q
  allowed_targets       = ["target1.example.com", "target2.example.com"]
}
`, relayName, caCert, relayHostname2),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_relay.test", "hostname", relayHostname2),
					helper.TestCheckResourceAttr("smallstep_relay.test", "allowed_targets.1", "target2.example.com"),
				),
			},
			{
				ResourceName:      "smallstep_relay.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
