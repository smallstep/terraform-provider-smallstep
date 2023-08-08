package collection_instance

import (
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/stretchr/testify/require"
)

func TestAccCollectionInstanceDataSource(t *testing.T) {
	t.Parallel()
	collection := utils.NewCollection(t)
	instance := utils.NewCollectionInstance(t, collection.Slug)
	config := fmt.Sprintf(`
data "smallstep_collection_instance" "test" {
	collection_slug = %q
	id = %q
}
`, collection.Slug, instance.Id)

	dataJSON, err := json.Marshal(instance.Data)
	require.NoError(t, err)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_collection_instance.test", "collection_slug", collection.Slug),
					helper.TestCheckResourceAttr("data.smallstep_collection_instance.test", "id", instance.Id),
					helper.TestCheckResourceAttr("data.smallstep_collection_instance.test", "data", string(dataJSON)),
					helper.TestMatchResourceAttr("data.smallstep_collection_instance.test", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestMatchResourceAttr("data.smallstep_collection_instance.test", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
		},
	})
}
