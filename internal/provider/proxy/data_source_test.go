package proxy

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccProxyDataSource(t *testing.T) {
	credential := utils.NewCredential(t)
	credentialID := *credential.Id
	proxyName := "tfprovider-" + utils.Slug(t)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(`
resource "smallstep_proxy" "test" {
  name             = %[1]q
  remote_address   = "proxy.example.com:3128"
  credentials      = [%[2]q]
  match_addresses  = ["*.example.com", "10.0.0.0/8"]
}

data "smallstep_proxy" "test" {
  id = smallstep_proxy.test.id
}
`, proxyName, credentialID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttrSet("data.smallstep_proxy.test", "id"),
					helper.TestCheckResourceAttr("data.smallstep_proxy.test", "name", proxyName),
					helper.TestCheckResourceAttr("data.smallstep_proxy.test", "remote_address", "proxy.example.com:3128"),
					helper.TestCheckResourceAttr("data.smallstep_proxy.test", "credentials.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_proxy.test", "credentials.0", credentialID),
					helper.TestCheckResourceAttr("data.smallstep_proxy.test", "match_addresses.#", "2"),
					helper.TestCheckResourceAttr("data.smallstep_proxy.test", "match_addresses.0", "*.example.com"),
					helper.TestCheckResourceAttr("data.smallstep_proxy.test", "match_addresses.1", "10.0.0.0/8"),
				),
			},
		},
	})
}
