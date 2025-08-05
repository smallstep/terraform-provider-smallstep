package strategy

import (
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const wifiConfig = `
	resource "smallstep_strategy" "wifi" {
		name = "WiFi Certificate"
		wifi = {
			ssid = "Big Corp Employee WiFi"
		}
		credential = {
			certificate_info = {
				x509 = {
					common_name = {
						device_metadata = "smallstep:identity"
					}
				}
				duration = "24h"
			}
		}
	}
`

func TestAccStrategyBrowser(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: wifiConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_strategy.wifi", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_strategy.wifi", "name", "WiFi Certificate"),
				),
			},
		},
	})
}
