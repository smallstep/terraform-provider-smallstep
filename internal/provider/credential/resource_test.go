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
	slug = %q
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
			}
			organization = {
				static = ["Example Inc"]
			}
		}
	}
	key = {
		type = "ECDSA_P384"
		protection = "HARDWARE_ATTESTED"
	}
	policy = {
		assurance = ["normal", "high"]
		os = ["Linux"]
		ownership = ["company"]
		source = ["Jamf"]
		tags = []
	}
	files = {
		root_file = "/var/ssl/ca.pem"
		crt_file = "/var/ssl/cert.pem"
		key_file = "/var/ssl/key.pem"
		uid = 501
		gid = 20
		mode = 256
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
		roo_file = ""
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
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.authority_id", authority.Id),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.duration", "168h"),

					// X509 fields
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.common_name.static", "My Device"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.common_name.device_metadata", "serial"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.static.#", "2"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.static.0", "staging.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.static.1", "*.staging.example.com"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.device_metadata.#", "2"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.device_metadata.0", "dns"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.sans.device_metadata.1", "email"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.organization.static.#", "1"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "certificate.x509.organization.static.0", "Example Inc"),

					// Key fields
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.type", "ECDSA_P384"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "key.protection", "HARDWARE_ATTESTED"),

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
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.uid", "501"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.gid", "20"),
					helper.TestCheckResourceAttr("smallstep_credential.test", "files.mode", "256"),
				),
			},
			{
				ResourceName:            "smallstep_credential.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate.duration"}, // 168h0m0s on import
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
