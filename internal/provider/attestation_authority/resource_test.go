package attestation_authority

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/collection"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
	"github.com/stretchr/testify/require"
)

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
		collection.NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccAttestationAuthorityResource(t *testing.T) {
	attestorRoot, attestorIntermediate := utils.CACerts(t)
	slug := utils.Slug(t)
	config := fmt.Sprintf(`
resource "smallstep_collection" "tpms" {
	slug = %q
}

resource "smallstep_attestation_authority" "aa" {
	name = %q
	attestor_roots = %q
	attestor_intermediates = %q
	depends_on = [smallstep_collection.tpms]
}
`, slug, slug, attestorRoot, attestorIntermediate)

	helper.Test(t, helper.TestCase{
		PreCheck: func() {
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_attestation_authority.aa", "name", slug),
					helper.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_roots", attestorRoot),
					helper.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_intermediates", attestorIntermediate),
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "root", regexp.MustCompile(`^-----BEGIN CERTIFICATE-----`)),
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "slug", regexp.MustCompile(`^[a-z1-9-]+$`)),
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				ResourceName:      "smallstep_attestation_authority.aa",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	nullIntermediate := fmt.Sprintf(`
resource "smallstep_collection" "tpms" {
	slug = %q
}

resource "smallstep_attestation_authority" "aa" {
	name = %q
	attestor_roots = %q
	depends_on = [smallstep_collection.tpms]
}
`, slug, slug, attestorRoot)

	helper.Test(t, helper.TestCase{
		PreCheck: func() {
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: nullIntermediate,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_roots", attestorRoot),
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "root", regexp.MustCompile(`^-----BEGIN CERTIFICATE-----`)),
				),
			},
		},
	})

	emptyIntermediate := fmt.Sprintf(`
resource "smallstep_collection" "tpms" {
	slug = %q
}

resource "smallstep_attestation_authority" "aa" {
	name = %q
	attestor_roots = %q
	attestor_intermediates = ""
	depends_on = [smallstep_collection.tpms]
}
`, slug, slug, attestorRoot)

	helper.Test(t, helper.TestCase{
		PreCheck: func() {
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyIntermediate,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_roots", attestorRoot),
					helper.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_intermediates", ""),
					helper.TestMatchResourceAttr("smallstep_attestation_authority.aa", "root", regexp.MustCompile(`^-----BEGIN CERTIFICATE-----`)),
				),
			},
		},
	})
}
