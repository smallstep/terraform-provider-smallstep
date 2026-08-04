package proxy

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

func TestAccProxyResource(t *testing.T) {
	credential := utils.NewCredential(t)
	credentialID := *credential.Id
	proxyName := "tfprovider-" + utils.Slug(t)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(`
resource "smallstep_proxy" "test" {
  name            = %[1]q
  remote_address  = "proxy.example.com:3128"
  credentials     = [%[2]q]
}
`, proxyName, credentialID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_proxy.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_proxy.test", "name", proxyName),
					helper.TestCheckResourceAttr("smallstep_proxy.test", "remote_address", "proxy.example.com:3128"),
					helper.TestCheckResourceAttr("smallstep_proxy.test", "credentials.0", credentialID),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "smallstep_proxy" "test" {
  name             = %[1]q
  remote_address   = "proxy2.example.com:8080"
  credentials      = [%[2]q]
  match_addresses  = ["*.example.com", "10.0.0.0/8"]
}
`, proxyName, credentialID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_proxy.test", "name", proxyName),
					helper.TestCheckResourceAttr("smallstep_proxy.test", "remote_address", "proxy2.example.com:8080"),
					helper.TestCheckResourceAttr("smallstep_proxy.test", "credentials.0", credentialID),
					helper.TestCheckResourceAttr("smallstep_proxy.test", "match_addresses.0", "*.example.com"),
					helper.TestCheckResourceAttr("smallstep_proxy.test", "match_addresses.1", "10.0.0.0/8"),
				),
			},
			{
				ResourceName:      "smallstep_proxy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
