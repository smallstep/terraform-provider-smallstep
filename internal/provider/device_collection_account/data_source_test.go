package device_collection_account

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccDeviceCollectionAccountDataSource(t *testing.T) {
	t.Parallel()
	dca, dcSlug := utils.NewDeviceCollectionAccount(t)

	config := fmt.Sprintf(`
data "smallstep_device_collection_account" "tester" {
	slug = %q
	device_collection_slug = %q
}
`, dca.Slug, dcSlug)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "slug", dca.Slug),
					resource.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "account_id", dca.AccountID),
					resource.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "device_collection_slug", dcSlug),
					resource.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "display_name", dca.DisplayName),
					resource.TestCheckResourceAttr("data.smallstep_device_collection_account.tester", "authority_id", *dca.AuthorityID),
				),
			},
		},
	})
}
