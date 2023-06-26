package provider

/*
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
	webhook := utils.New()
	config := fmt.Sprintf(`
data "smallstep_provisioner_webhook" "test" {
	authority_id = %q
	provisioner_id = %q
	id = %q
}
`, authority.Id, provisioner)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "id", authority.Id),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "name", authority.Name),
					resource.TestCheckResourceAttr("data.smallstep_authority.test", "kind", string(webhook.Kind)),
				),
			},
		},
	})
}
*/
