package identity_provider

import (
	"fmt"
	"regexp"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/stretchr/testify/require"
)

func TestAccIdpClient(t *testing.T) {
	require.NoError(t, sweep())

	ca, _ := utils.CACerts(t)

	const config = `
resource "smallstep_identity_provider" "my_idp" {
	trust_roots = %q
}

resource "smallstep_identity_provider_client" "my_idp_client" {
	redirect_uri = "https://example.com/callback"
	store_secret = true
	depends_on = [ smallstep_identity_provider.my_idp ]
}
`
	const config2 = `
resource "smallstep_identity_provider" "my_idp" {
	trust_roots = %q
}

resource "smallstep_identity_provider_client" "my_idp_client" {
	redirect_uri = "https://example.com/callback"
	depends_on = [ smallstep_identity_provider.my_idp ]
}
`
	const config3 = `
resource "smallstep_identity_provider" "my_idp" {
	trust_roots = %q
}

resource "smallstep_identity_provider_client" "my_idp_client" {
	redirect_uri = "https://example.com/callback2"
	depends_on = [ smallstep_identity_provider.my_idp ]
}
`
	const config4 = `
resource "smallstep_identity_provider" "my_idp" {
	trust_roots = %q
}

resource "smallstep_identity_provider_client" "my_idp_client" {
	redirect_uri = "https://example.com/callback"
	write_secret_file = "%s/my_idp_client_secret"
	depends_on = [ smallstep_identity_provider.my_idp ]
}
`
	cfg := fmt.Sprintf(config, ca)
	cfg2 := fmt.Sprintf(config2, ca)
	cfg3 := fmt.Sprintf(config3, ca)
	cfg4 := fmt.Sprintf(config4, ca, t.TempDir())
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: cfg,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_identity_provider_client.my_idp_client", "redirect_uri", "https://example.com/callback"),
					helper.TestMatchResourceAttr("smallstep_identity_provider_client.my_idp_client", "id", utils.UUIDRegexp),
					helper.TestMatchResourceAttr("smallstep_identity_provider_client.my_idp_client", "secret", regexp.MustCompile(`\w+`)),
					helper.TestCheckResourceAttr("smallstep_identity_provider_client.my_idp_client", "store_secret", "true"),
				),
			},
			{
				ResourceName:      "smallstep_identity_provider_client.my_idp_client",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: cfg2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckNoResourceAttr("smallstep_identity_provider_client.my_idp_client", "secret"),
				),
			},
			{
				Config: cfg3,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_identity_provider_client.my_idp_client", "redirect_uri", "https://example.com/callback2"),
					helper.TestCheckNoResourceAttr("smallstep_identity_provider_client.my_idp_client", "secret"),
				),
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: cfg2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckNoResourceAttr("smallstep_identity_provider_client.my_idp_client", "secret"),
				),
			},
			{
				ResourceName:      "smallstep_identity_provider_client.my_idp_client",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: cfg,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_identity_provider_client.my_idp_client", "secret", regexp.MustCompile(`\w+`)),
				),
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: cfg4,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_identity_provider_client.my_idp_client", "write_secret_file", regexp.MustCompile(`my_idp_client_secret$`)),
				),
			},
			{
				ResourceName:      "smallstep_identity_provider_client.my_idp_client",
				ImportState:       true,
				ImportStateVerify: false,
			},
			{
				Config: cfg2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckNoResourceAttr("smallstep_identity_provider_client.my_idp_client", "write_secret_file"),
				),
			},
		},
	})
}
