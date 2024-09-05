package device_collection_account

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccDeviceCollectionAccountDataSource(t *testing.T) {
	dca, dcSlug := utils.NewDeviceCollectionAccount(t)

	config := fmt.Sprintf(`
data "smallstep_device_collection_account" "tester" {
	slug = %q
	device_collection_slug = %q
}
`, dca.Slug, dcSlug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "slug", dca.Slug),
					helper.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "account_id", dca.AccountID),
					helper.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "device_collection_slug", dcSlug),
					helper.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "display_name", dca.DisplayName),
					helper.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "authority_id", dca.AuthorityID),
				),
			},
		},
	})
}
