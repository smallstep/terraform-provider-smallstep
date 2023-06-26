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
	name = "devices"
	kind = "ENRICHING"
	cert_type = "X509"
	server_type = "EXTERNAL"
	url = "https://example.com/hook"
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
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "name", "devices"),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "provisioner_id", *provisioner.Id),
					resource.TestCheckResourceAttr("smallstep_provisioner_webhook.external", "authority_id", authority.Id),
				),
			},
			/*
				{
					ResourceName:      "smallstep_provisioner.jwk",
					ImportState:       true,
					ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "eng@smallstep.com"),
					ImportStateVerify: false, // jwk serialized key may be different
				},
			*/
		},
	})
}
