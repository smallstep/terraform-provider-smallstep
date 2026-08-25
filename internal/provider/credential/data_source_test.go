package credential

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
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

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

func TestAccCredentialDataSource(t *testing.T) {
	authority := utils.NewAuthority(t)
	slug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	config := fmt.Sprintf(`
resource "smallstep_credential" "test" {
	slug = %q
	certificate = {
		authority_id = %q
		duration = "72h"
		x509 = {
			common_name = {
				static = "DataSource Test Device"
				device_metadata = "hostname"
				}
				sans = {
					static = ["device.example.com"]
					device_metadata = ["dns"]
			}
			organization = {
				static = ["Test Org"]
			}
			organizational_unit = {
				insecure_include_requested = true
			}
			given_name = {
				static = "Jane"
			}
			typed_sans = {
				email_addresses = {
					static = ["svc@example.com"]
				}
			}
			extended_key_usage = ["serverAuth"]
			custom_extensions = [{
				oid   = "1.3.6.1.4.1.44947"
				value = "dGVzdCBleHRlbnNpb24gdmFsdWU="
			}]
		}
		name_policy = {
			deny = {
				emails = ["@internal.example.com"]
			}
		}
	}
	key = {
		type       = "ECDSA_P384"
		protection = "NONE"
	}
	policy = {
		assurance = ["high"]
		os = ["Linux", "macOS"]
		ownership = ["company"]
		tags = ["test", "datasource"]
	}
	files = {
		root_file = "/opt/certs/ca.crt"
		crt_file = "/opt/certs/device.crt"
		key_file = "/opt/certs/device.key"
		key_format = "DEFAULT"
		uid = 2000
		gid = 2000
		mode = 420
	}
}

data "smallstep_credential" "test" {
	id = smallstep_credential.test.id
}
`, slug, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("data.smallstep_credential.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "slug", slug),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.authority_id", authority.Id),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.duration", "72h0m0s"),

					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.common_name.static", "DataSource Test Device"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.common_name.device_metadata", "hostname"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.sans.static.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.sans.static.0", "device.example.com"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.sans.device_metadata.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.sans.device_metadata.0", "dns"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.organization.static.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.organization.static.0", "Test Org"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.organizational_unit.insecure_include_requested", "true"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.given_name.static", "Jane"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.typed_sans.email_addresses.static.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.typed_sans.email_addresses.static.0", "svc@example.com"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.extended_key_usage.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.extended_key_usage.0", "serverAuth"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.custom_extensions.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.custom_extensions.0.oid", "1.3.6.1.4.1.44947"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.x509.custom_extensions.0.value", "dGVzdCBleHRlbnNpb24gdmFsdWU="),

					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.name_policy.deny.emails.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "certificate.name_policy.deny.emails.0", "@internal.example.com"),

					helper.TestCheckResourceAttr("data.smallstep_credential.test", "key.type", "ECDSA_P384"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "key.protection", "NONE"),

					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.assurance.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.assurance.0", "high"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.os.#", "2"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.os.0", "Linux"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.os.1", "macOS"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.ownership.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.ownership.0", "company"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.tags.#", "2"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.tags.0", "test"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "policy.tags.1", "datasource"),

					helper.TestCheckResourceAttr("data.smallstep_credential.test", "files.root_file", "/opt/certs/ca.crt"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "files.crt_file", "/opt/certs/device.crt"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "files.key_file", "/opt/certs/device.key"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "files.key_format", "DEFAULT"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "files.uid", "2000"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "files.gid", "2000"),
					helper.TestCheckResourceAttr("data.smallstep_credential.test", "files.mode", "420"),
				),
			},
		},
	})
}
