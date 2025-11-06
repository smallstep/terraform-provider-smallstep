package identity_provider

import (
	"fmt"
	"regexp"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/stretchr/testify/require"
)

func TestAccIdentityProviderResource(t *testing.T) {
	require.NoError(t, sweep())

	ca, _ := utils.CACerts(t)
	ca2, _ := utils.CACerts(t)

	const config = `
resource "smallstep_identity_provider" "my_idp" {
	trust_roots = %q
}`
	cfg := fmt.Sprintf(config, ca)
	cfg2 := fmt.Sprintf(config, ca2)

	issuerRx := regexp.MustCompile(`^https://.+.id.\w+`)
	authorizeRx := regexp.MustCompile(`^https.+/authorize$`)
	tokenRx := regexp.MustCompile(`^https.+/token$`)
	jwksRx := regexp.MustCompile(`^https.+/keys$`)
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: cfg,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_identity_provider.my_idp", "trust_roots", ca),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "issuer", issuerRx),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "authorize_endpoint", authorizeRx),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "token_endpoint", tokenRx),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "jwks_endpoint", jwksRx),
				),
			},
			{
				Config: cfg2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_identity_provider.my_idp", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_identity_provider.my_idp", "trust_roots", ca2),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "issuer", issuerRx),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "authorize_endpoint", authorizeRx),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "token_endpoint", tokenRx),
					helper.TestMatchResourceAttr("smallstep_identity_provider.my_idp", "jwks_endpoint", jwksRx),
				),
			},
			{
				ResourceName:                         "smallstep_identity_provider.my_idp",
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "issuer",
			},
		},
	})
}
