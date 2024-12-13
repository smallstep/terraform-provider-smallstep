package webhook

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/provisioner"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
		provisioner.NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccWebhookResource(t *testing.T) {
	t.Parallel()

	authority := utils.NewAuthority(t)
	provisioner, _ := utils.NewOIDCProvisioner(t, authority.Id)

	externalConfig := fmt.Sprintf(`
resource "smallstep_provisioner_webhook" "external" {
	authority_id = %q
	provisioner_id = %q
	name = "devices2"
	kind = "ENRICHING"
	cert_type = "X509"
	server_type = "EXTERNAL"
	url = "https://example.com/hook"
	bearer_token = "abc123"
}`, authority.Id, *provisioner.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: externalConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner_webhook.external", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "kind", "ENRICHING"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "name", "devices2"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "provisioner_id", *provisioner.Id),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "authority_id", authority.Id),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "url", "https://example.com/hook"),
					helper.TestMatchResourceAttr("smallstep_provisioner_webhook.external", "secret", regexp.MustCompile(`^[0-9A-Za-z+/]+={0,2}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "cert_type", "X509"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "server_type", "EXTERNAL"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "bearer_token", "abc123"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner_webhook.external",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s/%s", authority.Id, *provisioner.Id, "devices2"),
				ImportStateVerify: false, // secret and bearer token are never returned
			},
		},
	})

	basicAuthSSHConfig := fmt.Sprintf(`
resource "smallstep_provisioner_webhook" "basic" {
	authority_id = %q
	provisioner_id = %q
	name = "devices"
	kind = "ENRICHING"
	cert_type = "SSH"
	server_type = "EXTERNAL"
	url = "https://example.com/hook"
	basic_auth = {
		username = "user1"
		password = "pass1"
	}
}`, authority.Id, *provisioner.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: basicAuthSSHConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.basic", "basic_auth.username", "user1"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.basic", "basic_auth.password", "pass1"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.basic", "cert_type", "SSH"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner_webhook.basic",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s/%s", authority.Id, *provisioner.Id, "devices"),
				ImportStateVerify: false, // secret and basic_auth are never returned
			},
		},
	})

	hostedConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "agents" {
  authority_id = %q
  name = "Agents"
  type = "ACME_ATTESTATION"
  acme_attestation = {
    attestation_formats = ["apple"]
  }
}

resource "smallstep_provisioner_webhook" "hosted" {
	authority_id = smallstep_provisioner.agents.authority_id
	provisioner_id = smallstep_provisioner.agents.id
	name = "hosted"
	kind = "ENRICHING"
	cert_type = "X509"
	server_type = "HOSTED_ATTESTATION"
}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: hostedConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner_webhook.hosted", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.hosted", "kind", "ENRICHING"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.hosted", "name", "hosted"),
					helper.TestMatchResourceAttr("smallstep_provisioner_webhook.hosted", "provisioner_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.hosted", "authority_id", authority.Id),
					helper.TestMatchResourceAttr("smallstep_provisioner_webhook.hosted", "url", regexp.MustCompile(`/enrich/attested`)),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.hosted", "cert_type", "X509"),
					helper.TestCheckResourceAttr("smallstep_provisioner_webhook.hosted", "server_type", "HOSTED_ATTESTATION"),
				),
			},
		},
	})
}
