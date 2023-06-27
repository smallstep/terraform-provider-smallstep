package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: externalConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_provisioner_webhook.external", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "kind", "ENRICHING"),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "name", "devices2"),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "provisioner_id", *provisioner.Id),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "authority_id", authority.Id),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "url", "https://example.com/hook"),
					resource.TestMatchResourceAttr("smallstep_provisioner_webhook.external", "secret", regexp.MustCompile(`^[0-9A-Za-z+/]+={0,2}$`)),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "cert_type", "X509"),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "server_type", "EXTERNAL"),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "bearer_token", "abc123"),
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

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: basicAuthSSHConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.basic", "basic_auth.username", "user1"),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.basic", "basic_auth.password", "pass1"),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.basic", "cert_type", "SSH"),
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
}
