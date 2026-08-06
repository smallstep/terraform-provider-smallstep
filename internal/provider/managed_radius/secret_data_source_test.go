package managed_radius

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccManagedRadiusSecretDataSource(t *testing.T) {
	radius := utils.NewManagedRADIUS(t)

	config := fmt.Sprintf(`
data "smallstep_managed_radius_secret" "my_rad" {
	id = %q
}`, *radius.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_managed_radius_secret.my_rad", "id", *radius.Id),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius_secret.my_rad", "secret", *radius.Secret),
				),
			},
		},
	})
}
