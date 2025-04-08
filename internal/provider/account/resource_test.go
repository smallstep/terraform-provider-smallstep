package account

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
	},
	/*
		DataSourceFactories: []func() datasource.DataSource{
			NewDataSource,
		},
	*/
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

const minConfig = `
	resource "smallstep_account" "generic" {
		name = "Generic Client Certificate"
	}
`

const minConfigRename = `
	resource "smallstep_account" "generic" {
		name = "Access"
	}
`

const browserConfig = `
	resource "smallstep_account" "browser" {
		name = "Browser Certificate"
		browser = {}
	}
`

const wifiHostedRadiusConfig = `
resource "smallstep_account" "wifi_hosted_radius" {
	name = "WiFi Hosted Radius"
	wifi = {
		autojoin = true
		external_radius_server = false
		hidden = true
		network_access_server_ip = "0.0.0.1"
		ssid = "corpnet"
	}
}
`

const wifiHostedRadiusConfigUpdated = `
resource "smallstep_account" "wifi_hosted_radius" {
	name = "WiFi Hosted Radius"
	wifi = {
		autojoin = false
		external_radius_server = false
		network_access_server_ip = "0.0.0.2"
		ssid = "Corp Net"
	}
}
`

// key and reload have some required properties when they aren't null
// TODO can I handle them the same way I did x509?
const emptyConfig = `
	resource "smallstep_account" "generic" {
		name = "Generic Client Certificate"
		certificate = {}
		key = {
			format = "DEFAULT"
			type = "ECDSA_P256"
		}
		reload = {
			method = "AUTOMATIC"
		}
		policy = {}
	}
`

const fullX509Config = `
	resource "smallstep_account" "generic" {
		name = "Generic Client Certificate"
		certificate = {
			duration = "168h"
			crt_file = "db.crt"
			key_file = "db.key"
			root_file = "ca.crt"
			uid = 1001
			gid = 999
			mode = 256
			x509 = {
				common_name = {
					static = "Hello"
					device_metadata = "host"
				}
				sans = {
					static = ["user@example.com"]
					device_metadata = ["sans"]
				}
				organization = {
					static = ["static org"]
					device_metadata = ["org"]
				}
				organizational_unit = {
					static = ["static org unit"]
					device_metadata = ["ou"]
				}
				locality = {
					static = ["static loc"]
					device_metadata = ["locality"]
				}
				postal_code = {
					static = ["20252"]
					device_metadata = ["postal"]
				}
				country = {
					static = ["United States"]
					device_metadata = ["country"]
				}
				street_address = {
					static = ["1 Main"]
					device_metadata = ["street"]
				}
				province = {
					static = ["CA"]
					device_metadata = ["province"]
				}
			}
		}
		reload = {
			method = "SIGNAL"
			pid_file = "x.pid"
			signal = 1
		}
		key = {
			format = "DER"
			type = "ECDSA_P256"
			protection = "NONE"
		}
		policy = {
			assurance = ["high"]
			os = ["Windows", "macOS"]
			ownership = ["company"]
			source = ["Jamf", "Intune"]
			tags = ["mdm"]
		}
	}
`

const fullX509ConfigUpdated = `
	resource "smallstep_account" "generic" {
		name = "Asset"
		certificate = {
			type = "X509"
			duration = "24h"
			crt_file = "my.crt"
			key_file = "my.key"
			root_file = "my.crt"
			uid = 1002
			gid = 998
			mode = 6
			x509 = {
				common_name = {
					static = "Authorized Client"
					device_metadata = "email"
				}
				sans = {
					static = ["admin@example.com"]
					device_metadata = ["sans2"]
				}
				organization = {
					static = ["static org 2"]
					device_metadata = ["org2"]
				}
				organizational_unit = {
					static = ["static org unit 2"]
					device_metadata = ["ou2"]
				}
				locality = {
					static = ["static loc 2"]
					device_metadata = ["locality2"]
				}
				postal_code = {
					static = ["10"]
					device_metadata = ["postal", "postal2"]
				}
				country = {
					static = ["USA"]
					device_metadata = ["country2"]
				}
				street_address = {
					static = ["1 Main St."]
					device_metadata = ["street2"]
				}
				province = {
					static = ["California"]
					device_metadata = ["province2"]
				}
			}
		}
		reload = {
			method = "DBUS"
			unit_name = "protected.service"
		}
		key = {
			format = "DEFAULT"
			type = "ECDSA_P384"
			protection = "HARDWARE_ATTESTED"
		}
		policy = {
			assurance = ["normal"]
			os = ["Linux"]
			ownership = ["company", "user"]
			source = ["Jamf", "End-User"]
			tags = ["foo"]
		}
	}
`

const sshMinConfig = `
	resource "smallstep_account" "ssh" {
		name = "SSH User Certificate"
		certificate = {
			ssh = {
				key_id = {
					static = "123"
				}
			}
		}
	}
`

const sshFullConfig = `
	resource "smallstep_account" "ssh" {
		name = "SSH User Certificate"
		certificate = {
			ssh = {
				key_id = {
					device_metadata = "key"
				}
				principals = {
					static = ["eng"]
					device_metadata = ["role"]
				}
			}
		}
	}
`

func TestAccAccountSSH(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: sshMinConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.ssh", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "name", "SSH User Certificate"),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "certificate.ssh.key_id.static", "123"),
					helper.TestCheckNoResourceAttr("smallstep_account.ssh", "certificate.ssh.key_id.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.device_metadata"),
				),
			},
			{
				Config: sshFullConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.ssh", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.ssh", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "name", "SSH User Certificate"),
					helper.TestCheckNoResourceAttr("smallstep_account.ssh", "certificate.ssh.key_id.static"),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "certificate.ssh.key_id.device_metadata", "key"),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.static.0", "eng"),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.device_metadata.0", "role"),
				),
			},
			{
				Config: sshMinConfig,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.ssh", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.ssh", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "name", "SSH User Certificate"),
					helper.TestCheckResourceAttr("smallstep_account.ssh", "certificate.ssh.key_id.static", "123"),
					helper.TestCheckNoResourceAttr("smallstep_account.ssh", "certificate.ssh.key_id.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.ssh", "certificate.ssh.principals.device_metadata"),
				),
			},
		},
	})
}

func TestAccAccountRename(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.generic", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.generic", "name", "Generic Client Certificate"),
				),
			},
			{
				Config: minConfigRename,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.generic", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_account.generic", "name", "Access"),
				),
			},
			{
				ResourceName:      "smallstep_account.generic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAccountMinEmptyMin(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.generic", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.generic", "name", "Generic Client Certificate"),

					// default x509 fields
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.common_name.static", "Generic Client Certificate"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.common_name.device_metadata", "smallstep:identity"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.sans.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.sans.device_metadata.0", "smallstep:identity"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.sans.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.organization"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.locality"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.province"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.street_address"),

					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.postal_code"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.country"),

					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.crt_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.key_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.root_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.gid"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.uid"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.mode"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.ssh"),
				),
			},
			{
				Config: emptyConfig,
			},
			{
				Config: minConfig,
			},
		},
	})
}

func TestAccAccountEmptyFullEmpty(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.generic", "certificate.authority_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),

					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.crt_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.key_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.root_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.uid"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.gid"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.mode"),

					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.country.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.country.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.locality.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.locality.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.organization.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.organization.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.province.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.province.device_metadata"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.street_address.static"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.street_address.device_metadata"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "reload.method", "AUTOMATIC"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "reload.pid_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "reload.signal"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "reload.unit_file"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "key.format", "DEFAULT"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "key.type", "ECDSA_P256"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "key.protection"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "key.pub_file"),

					helper.TestCheckNoResourceAttr("smallstep_account.generic", "policy.assurance"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "policy.ownership"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "policy.os"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "policy.source"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "policy.tags"),
				),
			},
			{
				Config: fullX509Config,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.generic", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.generic", "certificate.authority_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.duration", "168h"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.crt_file", "db.crt"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.key_file", "db.key"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.root_file", "ca.crt"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.uid", "1001"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.gid", "999"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.mode", "256"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.common_name.static", "Hello"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.country.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.country.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.locality.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.locality.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organization.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organization.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.province.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.province.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.street_address.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.street_address.device_metadata.#", "1"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "reload.method", "SIGNAL"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "reload.pid_file", "x.pid"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "reload.signal", "1"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "key.format", "DER"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "key.type", "ECDSA_P256"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "key.protection", "NONE"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.assurance.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.assurance.0", "high"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.os.#", "2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.os.0", "Windows"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.os.1", "macOS"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.ownership.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.ownership.0", "company"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.source.#", "2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.source.0", "Jamf"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.source.1", "Intune"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.tags.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.tags.0", "mdm"),
				),
			},
			{
				ResourceName:            "smallstep_account.generic",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate.duration"}, // 168h0m0s on import
			},
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.generic", "certificate.authority_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.crt_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.crt_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.key_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.root_file"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.uid"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.gid"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.mode"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.street_address.static.#", "0"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "certificate.x509.street_address"),
				),
			},
		},
	})
}

func TestAccAccountX509FullUpdate(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fullX509Config,
			},
			{
				Config: fullX509ConfigUpdated,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.generic", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.generic", "certificate.authority_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.duration", "24h"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.crt_file", "my.crt"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.key_file", "my.key"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.root_file", "my.crt"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.uid", "1002"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.gid", "998"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.mode", "6"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.common_name.static", "Authorized Client"),
					// TODO add this above
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.common_name.device_metadata", "email"),

					// TODO above assert values not just length
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.country.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.country.static.0", "USA"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.country.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.country.device_metadata.0", "country2"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.locality.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.locality.static.0", "static loc 2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.locality.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.locality.device_metadata.0", "locality2"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organization.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organization.static.0", "static org 2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organization.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organization.device_metadata.0", "org2"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.static.0", "static org unit 2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.organizational_unit.device_metadata.0", "ou2"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.static.0", "10"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.device_metadata.#", "2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.device_metadata.0", "postal"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.postal_code.device_metadata.1", "postal2"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.province.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.province.static.0", "California"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.province.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.province.device_metadata.0", "province2"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.street_address.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.street_address.static.0", "1 Main St."),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.street_address.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "certificate.x509.street_address.device_metadata.0", "street2"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "reload.method", "DBUS"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "reload.unit_name", "protected.service"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "reload.signal"),
					helper.TestCheckNoResourceAttr("smallstep_account.generic", "reload.pid_file"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "key.format", "DEFAULT"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "key.type", "ECDSA_P384"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "key.protection", "HARDWARE_ATTESTED"),

					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.assurance.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.assurance.0", "normal"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.os.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.os.0", "Linux"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.ownership.#", "2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.ownership.0", "company"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.ownership.1", "user"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.source.#", "2"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.source.0", "Jamf"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.source.1", "End-User"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.tags.#", "1"),
					helper.TestCheckResourceAttr("smallstep_account.generic", "policy.tags.0", "foo"),
				),
			},
		},
	})
}

func TestAccAccountBrowser(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: browserConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.browser", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.browser", "name", "Browser Certificate"),
					helper.TestMatchResourceAttr("smallstep_account.browser", "certificate.authority_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_account.browser", "browser.%", "0"),
				),
			},
		},
	})
}

func TestAccAccountWiFiHostedRadius(t *testing.T) {
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: wifiHostedRadiusConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_account.wifi_hosted_radius", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "name", "WiFi Hosted Radius"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.autojoin", "true"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.external_radius_server", "false"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.hidden", "true"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.network_access_server_ip", "0.0.0.1"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ssid", "corpnet"),
					helper.TestMatchResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ca_chain", regexp.MustCompile("-----BEGIN CERTIFICATE-----")),
				),
			},
			{
				Config: wifiHostedRadiusConfigUpdated,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_account.wifi_hosted_radius", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.autojoin", "false"),
					helper.TestCheckNoResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.hidden"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.network_access_server_ip", "0.0.0.2"),
					helper.TestCheckResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ssid", "Corp Net"),
					helper.TestMatchResourceAttr("smallstep_account.wifi_hosted_radius", "wifi.ca_chain", regexp.MustCompile("-----BEGIN CERTIFICATE-----")),
				),
			},
		},
	})
}

func TestAccAccountWifiExternalRadius(t *testing.T) {
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
}

func TestAccAccountEthernet(t *testing.T) {
	root, _ := utils.CACerts(t)

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
