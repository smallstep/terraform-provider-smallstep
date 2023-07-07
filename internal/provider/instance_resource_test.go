package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccCollectionInstanceResource(t *testing.T) {
	t.Parallel()

	slug := utils.Slug(t)
	config := fmt.Sprintf(`
resource "smallstep_collection" "things" {
	slug = %q
}

resource "smallstep_collection_instance" "thing1" {
	collection_slug = %q
	id = "thing1"
	data = "{\"name\":\"thing1\"}"
	depends_on = [smallstep_collection.things]
}`, slug, slug)

	updated := fmt.Sprintf(`
resource "smallstep_collection" "things" {
	slug = %q
}

resource "smallstep_collection_instance" "thing1" {
	collection_slug = %q
	id = "thing1"
	data = "{\"name\":\"thing2\"}"
	depends_on = [smallstep_collection.things]
}`, slug, slug)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_collection_instance.thing1", "collection_slug", slug),
					resource.TestCheckResourceAttr("smallstep_collection_instance.thing1", "data", `{"name":"thing1"}`),
					resource.TestMatchResourceAttr("smallstep_collection_instance.thing1", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestMatchResourceAttr("smallstep_collection_instance.thing1", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				Config: updated,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_collection_instance.thing1", "collection_slug", slug),
					resource.TestCheckResourceAttr("smallstep_collection_instance.thing1", "data", `{"name":"thing2"}`),
					resource.TestMatchResourceAttr("smallstep_collection_instance.thing1", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestMatchResourceAttr("smallstep_collection_instance.thing1", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				ResourceName:  "smallstep_collection_instance.thing1",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/thing1", slug),
				// This only succeeds because the data JSON is simple enough to
				// have the same serialization every time. Change to `false` to
				// test more complex JSON.
				ImportStateVerify: true,
			},
		},
	})
}
