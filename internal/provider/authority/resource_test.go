package authority

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

func TestAccAuthorityResource(t *testing.T) {
	t.Parallel()

	devopsSlug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	devopsConfig := fmt.Sprintf(`
resource "smallstep_authority" "devops" {
	subdomain = "%s"
	name = "%s Authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}
`, devopsSlug, devopsSlug)

	caDomain := os.Getenv("SMALLSTEP_CA_DOMAIN")
	if caDomain == "" {
		caDomain = ".step-e2e.ca.smallstep.com"
	}

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: devopsConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_authority.devops", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_authority.devops", "domain", devopsSlug+caDomain),
					helper.TestMatchResourceAttr("smallstep_authority.devops", "fingerprint", regexp.MustCompile(`^[0-9a-z]{64}$`)),
					helper.TestMatchResourceAttr("smallstep_authority.devops", "root", regexp.MustCompile(`-{5}BEGIN`)),
					helper.TestMatchResourceAttr("smallstep_authority.devops", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				ResourceName:      "smallstep_authority.devops",
				ImportState:       true,
				ImportStateId:     devopsSlug + caDomain,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "smallstep_authority.devops",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	advancedSlug := "tfprovider-" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	advancedConfig := fmt.Sprintf(`
resource "smallstep_authority" "advanced" {
	subdomain = "%s"
	name = "%s Authority"
	type = "advanced"
	admin_emails = ["andrew@smallstep.com"]
	active_revocation = true
	intermediate_issuer = {
		name = "%s Intermediate"
		key_version = "RSA_SIGN_PKCS1_2048_SHA256"
		duration = "100h"
		max_path_length = 0
		name_constraints = {
			critical = true
			permitted_ip_ranges = ["10.32.0.0/12"]
			permitted_dns_domains = [".cluster.local"]
			permitted_email_addresses = ["eng@smallstep.com"]
			permitted_uri_domains = ["uri.cluster.local"]
		}
		subject = {
			common_name = "Issuer"
			country = "US"
			email_address = "test@smallstep.com"
			locality = "San Francisco"
			organization = "Engineering"
			organizational_unit = "Core"
			postal_code = "94108"
			province = "CA"
			serial_number = "1"
			street_address = "26 O'Farrell St"
		}
	}
	root_issuer = {
		name = "Root %s"
		key_version = "RSA_SIGN_PKCS1_2048_SHA256"
		duration = "1000h"
		max_path_length = "1"
		name_constraints = {
			critical = false
			excluded_ip_ranges = ["10.96.0.0/12"]
			excluded_dns_domains = ["example.com"]
			excluded_email_addresses = ["eng@example.com"]
			excluded_uri_domains = ["uri:example.com"]
		}
	}
}
`, advancedSlug, advancedSlug, advancedSlug, advancedSlug)
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: advancedConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_authority.advanced", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_authority.advanced", "domain", advancedSlug+caDomain),
					helper.TestMatchResourceAttr("smallstep_authority.advanced", "fingerprint", regexp.MustCompile(`^[0-9a-z]{64}$`)),
					helper.TestMatchResourceAttr("smallstep_authority.advanced", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
