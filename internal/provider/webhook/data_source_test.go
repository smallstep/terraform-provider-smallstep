package webhook

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccWebhookDataSource(t *testing.T) {
	t.Parallel()
	authority := utils.NewAuthority(t)
	provisioner, _ := utils.NewOIDCProvisioner(t, authority.Id)
	webhook := utils.NewWebhook(t, *provisioner.Id, authority.Id)
	config := fmt.Sprintf(`
data "smallstep_provisioner_webhook" "test" {
	authority_id = %q
	provisioner_id = %q
	id = %q
}
`, authority.Id, *provisioner.Id, *webhook.Id)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_provisioner_webhook.test", "id", *webhook.Id),
					resource.TestCheckResourceAttr("data.smallstep_provisioner_webhook.test", "name", webhook.Name),
					resource.TestCheckResourceAttr("data.smallstep_provisioner_webhook.test", "kind", string(webhook.Kind)),
					resource.TestCheckResourceAttr("data.smallstep_provisioner_webhook.test", "cert_type", string(webhook.CertType)),
					resource.TestCheckResourceAttr("data.smallstep_provisioner_webhook.test", "server_type", string(webhook.ServerType)),
					resource.TestCheckResourceAttr("data.smallstep_provisioner_webhook.test", "url", *webhook.Url),
				),
			},
		},
	})
}
