package attestation_authority

import (
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func init() {
	resource.AddTestSweepers("smallstep_attestation_authority", &resource.Sweeper{
		Name: "smallstep_attestation_authority",
		F: func(region string) error {
			return utils.SweepAttestationAuthorities()
		},
	})
}
