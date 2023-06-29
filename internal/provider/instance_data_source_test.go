package provider

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/stretchr/testify/require"
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

	dataJSON, err := json.Marshal(instance.Data)
	require.NoError(t, err)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_collection_instance.test", "collection_slug", collection.Slug),
					resource.TestCheckResourceAttr("data.smallstep_collection_instance.test", "id", instance.Id),
					resource.TestCheckResourceAttr("data.smallstep_collection_instance.test", "data", string(dataJSON)),
					resource.TestMatchResourceAttr("data.smallstep_collection_instance.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestMatchResourceAttr("data.smallstep_collection_instance.test", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
