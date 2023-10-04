package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/stretchr/testify/require"
)

func init() {
	resource.AddTestSweepers("smallstep_attestation_authority", &resource.Sweeper{
		Name: "smallstep_attestation_authority",
		F: func(region string) error {
			return utils.SweepAttestationAuthorities()
		},
	})
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_attestation_authority.aa", "name", slug)
					resource.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_roots", attestorRoot),
					resource.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_intermediates", attestorIntermediate),
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "root", regexp.MustCompile(`^-----BEGIN CERTIFICATE-----`)),
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "slug", regexp.MustCompile(`^[a-z1-9-]+$`)),
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: nullIntermediate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_roots", attestorRoot),
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "root", regexp.MustCompile(`^-----BEGIN CERTIFICATE-----`)),
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

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: emptyIntermediate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_roots", attestorRoot),
					resource.TestCheckResourceAttr("smallstep_attestation_authority.aa", "attestor_intermediates", ""),
					resource.TestMatchResourceAttr("smallstep_attestation_authority.aa", "root", regexp.MustCompile(`^-----BEGIN CERTIFICATE-----`)),
				),
			},
		},
	})
}
