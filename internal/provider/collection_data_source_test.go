package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccCollectionDataSource(t *testing.T) {
	t.Parallel()
	collection := utils.NewCollection(t)
	config := fmt.Sprintf(`
data "smallstep_collection" "test" {
	slug = %q
}
`, collection.Slug)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_collection.test", "slug", collection.Slug),
					resource.TestCheckResourceAttr("data.smallstep_collection.test", "instance_count", "0"),
					resource.TestCheckResourceAttr("data.smallstep_collection.test", "display_name", collection.DisplayName),
					resource.TestMatchResourceAttr("data.smallstep_collection.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestMatchResourceAttr("data.smallstep_collection.test", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
