package account

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccAccountDataSource(t *testing.T) {
	account := utils.NewAccount(t)

	config := fmt.Sprintf(`
data "smallstep_account" "generic" {
	id = %q
}`, account.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_account.generic", "id", account.Id),
					helper.TestCheckResourceAttr("data.smallstep_account.generic", "name", account.Name),
				),
			},
		},
	})
}
