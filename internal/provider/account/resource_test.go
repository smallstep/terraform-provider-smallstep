package account

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

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

func TestAccWorkloadResource(t *testing.T) {
	t.Parallel()
	const browsers = `
resource "smallstep_account" "browsers" {
	name = "Browsers"
	browser = {}
}
`
	const browsers2 = `
resource "smallstep_account" "browsers" {
	name = "Browsers2"
	browser = {}
}
`

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: browsers,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.browsers", plancheck.ResourceActionCreate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.browsers", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_account.browsers", "name", "Browsers"),
				),
			},
			{
				Config: browsers2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.browsers", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.browsers", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_account.browsers", "name", "Browsers2"),
				),
			},
			{
				ResourceName:      "smallstep_account.browsers",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	const wifiHostedRadius = `
resource "smallstep_account" "wifi_hosted_radius" {
	name = "WiFi Hosted Radius"
	wifi = {
		autojoin = true
		external_radius_server = false
		hidden = true
		network_access_server_ip = "1.2.3.4"
		ssid = "corpnet"
	}
}
`
	const wifiHostedRadius2 = `
resource "smallstep_account" "wifi_hosted_radius" {
	name = "WiFi Hosted Radius"
	wifi = {
		autojoin = false
		external_radius_server = false
		network_access_server_ip = "5.6.7.8"
		ssid = "Corp Net"
	}
}
`
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: wifiHostedRadius,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.wifi_hosted_radius", plancheck.ResourceActionCreate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.wifi_hosted_radius", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "name", "WiFi Hosted Radius"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.external_radius_server", "false"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.hidden", "true"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.network_access_server_ip", "1.2.3.4"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ssid", "corpnet"),
					helper.TestMatchResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ca_chain", regexp.MustCompile("-----BEGIN CERTIFICATE-----")),
				),
			},
			{
				Config: wifiHostedRadius2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.wifi_hosted_radius", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.autojoin", "false"),
					helper.TestCheckNoResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.hidden"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.network_access_server_ip", "5.6.7.8"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ssid", "Corp Net"),
					helper.TestMatchResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ca_chain", regexp.MustCompile("-----BEGIN CERTIFICATE-----")),
				),
			},
		},
	})

	root, _ := utils.CACerts(t)
	wifiExternalRadius := fmt.Sprintf(`
resource "smallstep_account" "wifi_byoradius" {
	name = "WiFi Hosted Radius"
	wifi = {
		external_radius_server = true
		ca_chain = %q
		network_access_server_ip = "1.2.3.4"
		ssid = "corpnet"
	}
}
`, root)
	root2, _ := utils.CACerts(t)
	wifiExternalRadius2 := fmt.Sprintf(`
resource "smallstep_account" "wifi_byoradius" {
	name = "WiFi Hosted Radius"
	wifi = {
		external_radius_server = true
		ca_chain = %q
		network_access_server_ip = "1.2.3.4"
		ssid = "corpnet"
		hidden = true
		autojoin = true
	}
}
`, root2)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: wifiExternalRadius,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.wifi_byoradius", plancheck.ResourceActionCreate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.wifi_byoradius", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "name", "WiFi Hosted Radius"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "wifi.external_radius_server", "true"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "wifi.network_access_server_ip", "1.2.3.4"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "wifi.ca_chain", root),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "wifi.ssid", "corpnet"),
					helper.TestCheckNoResourceAttr("smallstep_account.wifi_byoradius", "wifi.autojoin"),
					helper.TestCheckNoResourceAttr("smallstep_account.wifi_byoradius", "wifi.hidden"),
				),
			},
			{
				Config: wifiExternalRadius2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.wifi_byoradius", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.wifi_byoradius", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "wifi.ca_chain", root2),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "wifi.hidden", "true"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_byoradius", "wifi.autojoin", "true"),
				),
			},
		},
	})

	const vpn = `
resource "smallstep_account" "vpn" {
	name = "VPN"
	vpn = {
		autojoin = true
		connection_type = "IPSec"
		remote_address = "vpn.example.com"
		vendor = "F5"
	}
}
`
	vpn2 := fmt.Sprintf(`
resource "smallstep_account" "vpn" {
	name = "VPN"
	vpn = {
		autojoin = false
		connection_type = "IKEv2"
		remote_address = "ike.example.com"
		ike = {
			ca_chain = %q
			eap = true
			remote_id = "foo"
		}
	}
}
`, root)

	const vpn3 = `
resource "smallstep_account" "vpn" {
	name = "VPN"
	vpn = {
		connection_type = "SSL"
		remote_address = "ssl.example.com"
		vendor = "Cisco"
	}
}
`

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: vpn,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.vpn", plancheck.ResourceActionCreate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.connection_type", "IPSec"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.remote_address", "vpn.example.com"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.vendor", "F5"),
					helper.TestCheckNoResourceAttr("smallstep_account.vpn", "vpn.ike"),
				),
			},
			{
				Config: vpn2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.vpn", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.autojoin", "false"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.connection_type", "IKEv2"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.remote_address", "ike.example.com"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.ike.ca_chain", root),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.ike.eap", "true"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.ike.remote_id", "foo"),
				),
			},
			{
				Config: vpn3,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.vpn", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.connection_type", "SSL"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.remote_address", "ssl.example.com"),
					helper.TestCheckResourceAttr("smallstep_account.vpn", "vpn.vendor", "Cisco"),
					helper.TestCheckNoResourceAttr("smallstep_account.vpn", "vpn.ike"),
					helper.TestCheckNoResourceAttr("smallstep_account.vpn", "vpn.autojoin"),
				),
			},
		},
	})

	const ethernet1 = `
resource "smallstep_account" "ethernet" {
	name = "Ethernet Hosted Radius"
	ethernet = {
		autojoin = true
		external_radius_server = false
		network_access_server_ip = "1.2.3.4"
	}
}
`
	ethernet2 := fmt.Sprintf(`
resource "smallstep_account" "ethernet" {
	name = "Ethernet BYORadius"
	ethernet = {
		autojoin = false
		external_radius_server = true
		ca_chain = %q
	}
}
`, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: ethernet1,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.ethernet", plancheck.ResourceActionCreate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.ethernet", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_account.ethernet", "name", "Ethernet Hosted Radius"),
					helper.TestCheckResourceAttr("smallstep_account.ethernet", "ethernet.autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_account.ethernet", "ethernet.external_radius_server", "false"),
					helper.TestCheckResourceAttr("smallstep_account.ethernet", "ethernet.network_access_server_ip", "1.2.3.4"),
					helper.TestMatchResourceAttr("smallstep_account.ethernet", "ethernet.ca_chain", regexp.MustCompile("-----BEGIN CERTIFICATE-----")),
				),
			},
			{
				Config: ethernet2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.ethernet", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_account.ethernet", "ethernet.autojoin", "false"),
					helper.TestCheckNoResourceAttr("smallstep_account.ethernet", "ethernet.network_access_server_ip"),
					helper.TestCheckResourceAttr("smallstep_account.ethernet", "ethernet.ca_chain", root),
					helper.TestCheckResourceAttr("smallstep_account.ethernet", "ethernet.external_radius_server", "true"),
				),
			},
		},
	})
}
