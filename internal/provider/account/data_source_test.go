package account

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccAccountDataSource(t *testing.T) {
	t.Parallel()
	account, wifi := utils.NewAccount(t)
	config := fmt.Sprintf(`
data "smallstep_account" "wifi" {
	id = %q
}
`, *account.Id)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_account.wifi", "id", *account.Id),
					resource.TestCheckResourceAttr("data.smallstep_account.wifi", "name", account.Name),
					resource.TestCheckResourceAttr("data.smallstep_account.wifi", "wifi.ssid", wifi.Ssid),
					resource.TestCheckResourceAttr("data.smallstep_account.wifi", "wifi.network_access_server_ip", *wifi.NetworkAccessServerIP),
					resource.TestMatchResourceAttr("data.smallstep_account.wifi", "wifi.ca_chain", regexp.MustCompile("-----BEGIN CERTIFICATE-----")),
				),
			},
		},
	})
}
