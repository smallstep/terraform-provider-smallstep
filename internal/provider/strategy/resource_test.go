package strategy

import (
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const wifiConfig = `resource "smallstep_strategy" "wifi" {
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
}`

func TestAccStrategyWifi(t *testing.T) {
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

const browserConfig = `
	resource "smallstep_strategy" "browser" {
		name = "Browser Certificate"
		browser = {
			match_addresses = ["https://example.com"]
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
				Config: browserConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_strategy.browser", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_strategy.browser", "name", "Browser Certificate"),
				),
			},
		},
	})
}

func TestStrategSSH(t *testing.T) {
	const sshConfig = `resource "smallstep_strategy" "ssh" {
	name = "SSH Certificate"
	ssh = {}
	credential = {
		certificate_info = {
			ssh = {
				key_id = {
					device_metadata = "smallstep:identity"
				}
				principals = {
					device_metadata = ["SSH.Principals", "smallstep:identity"]
				}
			}
		}
	}
}`
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: sshConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_strategy.ssh", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_strategy.ssh", "name", "SSH Certificate"),
				),
			},
		},
	})
}

func TestAccStrategRelay(t *testing.T) {
	const relayConfig = `resource "smallstep_strategy" "relay" {
	name = "Relay Certificate"
	relay = {
		match_domains: ["example.com"]
		regions: ["US_CENTRAL1"]
	}
	credential = {
		certificate_info = {
			x509 = {
				common_name = {
					device_metadata = "smallstep:identity"
				}
			}
		}
	}
}`
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: relayConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_strategy.ssh", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_strategy.ssh", "name", "Relay Certificate"),
				),
			},
		},
	})
}
