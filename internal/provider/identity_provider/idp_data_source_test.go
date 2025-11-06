package identity_provider

import (
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccIdentityProviderDataSource(t *testing.T) {
	idp := utils.NewIdentityProvider(t)

	const config = `data "smallstep_identity_provider" "my_idp" {}`

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_identity_provider.my_idp", "trust_roots", idp.TrustRoots),
					helper.TestCheckResourceAttr("data.smallstep_identity_provider.my_idp", "issuer", *idp.Issuer),
					helper.TestCheckResourceAttr("data.smallstep_identity_provider.my_idp", "jwks_endpoint", *idp.JwksEndpoint),
					helper.TestCheckResourceAttr("data.smallstep_identity_provider.my_idp", "token_endpoint", *idp.TokenEndpoint),
					helper.TestCheckResourceAttr("data.smallstep_identity_provider.my_idp", "authorize_endpoint", *idp.AuthorizeEndpoint),
				),
			},
		},
	})
}
