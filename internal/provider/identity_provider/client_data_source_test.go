package identity_provider

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccIdentityProviderClientDataSource(t *testing.T) {
	utils.NewIdentityProvider(t)
	client := utils.NewIdentityProviderClient(t)

	const config = `
data "smallstep_identity_provider_client" "my_idp_client" {
	id = %q
}`
	cfg := fmt.Sprintf(config, *client.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: cfg,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_identity_provider_client.my_idp_client", "id", *client.Id),
					helper.TestCheckResourceAttr("data.smallstep_identity_provider_client.my_idp_client", "redirect_uri", client.RedirectURI),
				),
			},
		},
	})
}
