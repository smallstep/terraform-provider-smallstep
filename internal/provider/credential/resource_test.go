package credential

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccCredentialResource(t *testing.T) {
	authority := utils.NewAuthority(t)
	slug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)

	minConfig := fmt.Sprintf(`
resource "smallstep_credential" "test" {
	slug = %q
	certificate = {
		authority_id = %q
		x509 = {
			common_name = {
				static = "Test Device"
			}
		}
	}
	key = {
		type = "ECDSA_P256"
		protection = "HARDWARE"
	}
}
`, slug, authority.Id)
	fullConfig := fmt.Sprintf(`
resource "smallstep_credential" "test" {
	slug            = %q
	management_mode = "agent"
	certificate = {
		authority_id = %q
		duration = "168h"
		x509 = {
			common_name = {
				static = "My Device"
				device_metadata = "serial"
			}
			sans = {
				static = ["staging.example.com", "*.staging.example.com"]
				device_metadata = ["dns", "email"]
				insecure_include_requested = true
			}
			organization = {
				static = ["Example Inc"]
			}
			organizational_unit = {
				static          = ["Engineering"]
				device_metadata = ["department"]
			}
			locality = {
				static          = ["San Francisco"]
				device_metadata = ["city"]
			}
			province = {
				static          = ["California"]
				device_metadata = ["state"]
			}
			street_address = {
				static          = ["123 Main St"]
				device_metadata = ["address"]
			}
			postal_code = {
				static          = ["94105"]
				device_metadata = ["zip"]
			}
			country = {
				static          = ["US"]
				device_metadata = ["country"]
			}
			given_name = {
				static = "Jane"
			}
			serial_number = {
				static = "12345"
			}
			surname = {
				static = "Doe"
			}
			typed_sans = {
				dns_names = {
					static = ["svc.example.com"]
				}
				ip_addresses = {
					static = ["10.0.0.1"]
				}
				email_addresses = {
					static = ["svc@example.com"]
				}
				uris = {
					static = ["https://svc.example.com"]
				}
				user_principal_names = {
					static = ["svc@example.com"]
				}
			}
			extended_key_usage = ["serverAuth", "clientAuth"]
			custom_extensions = [{
				oid      = "1.3.6.1.4.1.44947"
				critical = false
				value    = "dGVzdCBleHRlbnNpb24gdmFsdWU="
			}]
		}
		name_policy = {
			allow = {
				common_names = ["Allowed CN"]
				dns          = ["*.example.com"]
				emails       = ["allowed@example.com"]
				ips          = ["10.0.0.0/8"]
				uris         = ["allowed.example.com"]
			}
			deny = {
				common_names = ["Denied CN"]
				dns          = ["*.denied.example.com"]
				emails       = ["denied@example.com"]
				ips          = ["192.168.0.0/16"]
				uris         = ["denied.example.com"]
			}
			allow_wildcard_names = true
		}
	}
	key = {
		type          = "ECDSA_P384"
		protection    = "HARDWARE_ATTESTED"
		compatibility = "DEFAULT"
		store         = "DEFAULT"
	}
	policy = {
		assurance = ["normal", "high"]
		os = ["Linux"]
		ownership = ["company"]
		source = ["Jamf"]
		tags = []
	}
	files = {
		root_file  = "/var/ssl/ca.pem"
		crt_file   = "/var/ssl/cert.pem"
		key_file   = "/var/ssl/key.pem"
		key_format = "DEFAULT"
		uid        = 501
		gid        = 20
		mode       = 256
	}
}
`, slug, authority.Id)

	emptyConfig := fmt.Sprintf(`
resource "smallstep_credential" "test" {
	slug = %q
	certificate = {
		authority_id = %q
		x509 = {
			common_name = {
				device_metadata = "smallstep:identity"
			}
		}
	}
	key = {
		type = "ECDSA_P384"
		protection = "HARDWARE_ATTESTED"
	}
	policy = {}
	files = {}
}
`, slug, authority.Id)

	emptyConfig2 := fmt.Sprintf(`
resource "smallstep_credential" "test" {
	slug = %q
	certificate = {
		authority_id = %q
		x509 = {
			common_name = {
				device_metadata = "smallstep:identity"
			}
		}
	}
	key = {
		type = "ECDSA_P384"
		protection = "HARDWARE_ATTESTED"
	}
	policy = {
		os = []
		assurance = []
	}
	files = {
		crt_file = ""
		key_file = ""
		root_file = ""
	}
}
`, slug, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_credential.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_credential.test", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.authority_id", authority.Id),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.common_name.static", "Test Device"),
					helper.TestCheckNoResourceAttr("smallstep_credential.test", "certificate.x509.common_name.device_metadata"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.type", "ECDSA_P256"),
					helper.TestCheckNoResourceAttr("smallstep_credential.test", "key.pub_file"),
					helper.TestCheckNoResourceAttr("smallstep_credential.test", "policy"),
					helper.TestCheckNoResourceAttr("smallstep_credential.test", "files"),
				),
			},
			{
				ResourceName:      "smallstep_credential.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_credential.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_credential.test", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.authority_id", authority.Id),
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.type", "ECDSA_P384"),
				),
			},
			{
				Config: minConfig,
			},
			{
				Config: emptyConfig,
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_credential.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_credential.test", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_credential.test", "management_mode", "agent"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.authority_id", authority.Id),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.duration", "168h"),

					// X509 fields
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.common_name.static", "My Device"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.common_name.device_metadata", "serial"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.insecure_include_requested", "true"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.static.#", "2"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.static.0", "staging.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.static.1", "*.staging.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.device_metadata.#", "2"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.device_metadata.0", "dns"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.device_metadata.1", "email"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.organization.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.organization.static.0", "Example Inc"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.organizational_unit.static.0", "Engineering"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.organizational_unit.device_metadata.0", "department"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.locality.static.0", "San Francisco"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.locality.device_metadata.0", "city"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.province.static.0", "California"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.province.device_metadata.0", "state"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.street_address.static.0", "123 Main St"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.street_address.device_metadata.0", "address"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.postal_code.static.0", "94105"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.postal_code.device_metadata.0", "zip"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.country.static.0", "US"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.country.device_metadata.0", "country"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.given_name.static", "Jane"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.serial_number.static", "12345"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.surname.static", "Doe"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.dns_names.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.dns_names.static.0", "svc.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.ip_addresses.static.0", "10.0.0.1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.email_addresses.static.0", "svc@example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.uris.static.0", "https://svc.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.user_principal_names.static.0", "svc@example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.extended_key_usage.#", "2"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.extended_key_usage.0", "serverAuth"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.extended_key_usage.1", "clientAuth"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.custom_extensions.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.custom_extensions.0.oid", "1.3.6.1.4.1.44947"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.custom_extensions.0.critical", "false"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.custom_extensions.0.value", "dGVzdCBleHRlbnNpb24gdmFsdWU="),

					// Name policy
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow.common_names.0", "Allowed CN"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow.dns.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow.dns.0", "*.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow.emails.0", "allowed@example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow.ips.0", "10.0.0.0/8"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow.uris.0", "allowed.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.deny.common_names.0", "Denied CN"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.deny.dns.0", "*.denied.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.deny.emails.0", "denied@example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.deny.ips.0", "192.168.0.0/16"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.deny.uris.0", "denied.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow_wildcard_names", "true"),

					// Key fields
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.type", "ECDSA_P384"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.protection", "HARDWARE_ATTESTED"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.compatibility", "DEFAULT"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.store", "DEFAULT"),

					// Policy fields
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.assurance.#", "2"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.assurance.0", "normal"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.assurance.1", "high"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.os.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.os.0", "Linux"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.ownership.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.ownership.0", "company"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.source.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.source.0", "Jamf"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "policy.tags.#", "0"),

					// Files fields
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.root_file", "/var/ssl/ca.pem"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.crt_file", "/var/ssl/cert.pem"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.key_file", "/var/ssl/key.pem"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.key_format", "DEFAULT"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.uid", "501"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.gid", "20"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.mode", "256"),
				),
			},
			{
				ResourceName:            "smallstep_credential.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{
					"certificate.duration", // 168h0m0s on import
					"key.store",            // DEFAULT is dropped by the API and unrecoverable from a fresh import
				},
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_credential.test", "id", utils.UUIDRegexp),
				),
			},
			{
				ResourceName: "smallstep_credential.test",
				ImportState:  true,
			},
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_credential.test", "id", utils.UUIDRegexp),
				),
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyConfig2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_credential.test", "id", utils.UUIDRegexp),
				),
			},
			{
				ResourceName: "smallstep_credential.test",
				ImportState:  true,
			},
			{
				Config: emptyConfig,
			},
		},
	})
}

// TestCredentialResourceOptionalFieldEmptyValues exercises the equivalent empty value
// for optional fields
func TestCredentialResourceOptionalFieldEmptyValues(t *testing.T) {
	authority := utils.NewAuthority(t)
	{
		slug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
		config := fmt.Sprintf(`
resource "smallstep_credential" "test" {
	slug = %q
	certificate = {
		authority_id = %q
		x509 = {
			common_name = {
				static = "Test Device"
			}
			sans = {
				static                      = ["device.example.com"]
				insecure_include_requested = false
			}
			custom_extensions = [{
				oid      = "1.3.6.1.4.1.44947"
				value    = "dGVzdA=="
				critical = false
			}]
			organizational_unit = {
				static = ["eng"]
				device_metadata = []
			}
			typed_sans = {
				email_addresses = {
					static = ["svc@example.com"]
				}
				dns_names = {
					static = []
				}
			}
			extended_key_usage = []
		}
		name_policy = {
			allow = {
				dns = ["*.example.com"]
				common_names = ["My Common Name"]
			}
			allow_wildcard_names = false
		}
	}
	key = {
		type       = "ECDSA_P256"
		protection = "NONE"
	}
	policy = {
		os = ["Linux"]
		assurance = []
	}
}
`, slug, authority.Id)
		helper.Test(t, helper.TestCase{
			ProtoV6ProviderFactories: providerFactories,
			Steps: []helper.TestStep{
				{
					Config: config,
					Check: helper.ComposeAggregateTestCheckFunc(
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.static.0", "device.example.com"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.insecure_include_requested", "false"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow_wildcard_names", "false"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.name_policy.allow.common_names.0", "My Common Name"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.custom_extensions.0.critical", "false"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.extended_key_usage.#", "0"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "policy.os.0", "Linux"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "policy.assurance.#", "0"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.email_addresses.static.0", "svc@example.com"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.typed_sans.dns_names.static.#", "0"),
						helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.organizational_unit.device_metadata.#", "0"),
					),
				},
			},
		})
	}
}
