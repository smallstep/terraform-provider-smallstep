package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccCollectionInstanceDataSource(t *testing.T) {
	t.Parallel()
	collection := utils.NewCollection(t)
	instance := utils.NewInstance(t, collection.Slug)
	config := fmt.Sprintf(`
data "smallstep_collection_instance" "test" {
	collection_slug = %q
	id = %q
}
`, collection.Slug, instance.Id)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_collection_instance.test", "collection_slug", collection.Slug),
					// TODO data
					// TODO ID
					resource.TestMatchResourceAttr("data.smallstep_collection_instance.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestMatchResourceAttr("data.smallstep_collection_instance.test", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
