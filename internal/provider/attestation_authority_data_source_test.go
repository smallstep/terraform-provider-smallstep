package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccAttestationAuthorityDataSource(t *testing.T) {
	// Don't parallelize because of limit of 1 attestation authority per team

	collection := utils.NewCollection(t)
	// Catalog may not be the slug passed in if the attestation authority already existed
	aa := utils.FixAttestationAuthority(t, collection.Slug)

	config := fmt.Sprintf(`
data "smallstep_attestation_authority" "test" {
	id = %q
}
`, *aa.Id)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_attestation_authority.test", "id", *aa.Id),
					resource.TestCheckResourceAttr("data.smallstep_attestation_authority.test", "name", aa.Name),
					resource.TestCheckResourceAttr("data.smallstep_attestation_authority.test", "catalog", aa.Catalog),
					resource.TestCheckResourceAttr("data.smallstep_attestation_authority.test", "slug", *aa.Slug),
					resource.TestCheckResourceAttr("data.smallstep_attestation_authority.test", "root", *aa.Root),
					resource.TestCheckResourceAttr("data.smallstep_attestation_authority.test", "attestor_roots", aa.AttestorRoots),
					resource.TestCheckResourceAttr("data.smallstep_attestation_authority.test", "attestor_intermediates", *aa.AttestorIntermediates),
					resource.TestMatchResourceAttr("data.smallstep_attestation_authority.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
