package sso_integration

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccSsoIntegrationDataSource(t *testing.T) {
	utils.NewIdentityProvider(t)
	redirectURI := "https://tfprovider.example.com/oauth/callback"
	failureURI := "https://app.example.com/oauth/error"

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(`
resource "smallstep_sso_integration" "test" {
  redirect_uri             = %[1]q
  lifecycle_failure_uri    = %[2]q
}

data "smallstep_sso_integration" "test" {
  id = smallstep_sso_integration.test.id
}
`, redirectURI, failureURI),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttrSet("data.smallstep_sso_integration.test", "id"),
					helper.TestCheckResourceAttr("data.smallstep_sso_integration.test", "redirect_uri", redirectURI),
					helper.TestCheckResourceAttr("data.smallstep_sso_integration.test", "lifecycle_failure_uri", failureURI),
				),
			},
		},
	})
}
