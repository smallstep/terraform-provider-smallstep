package collection_instance

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/collection"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		collection.NewResource,
		NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

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

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_collection_instance.thing1", "collection_slug", slug),
					helper.TestCheckResourceAttr("smallstep_collection_instance.thing1", "data", `{"name":"thing1"}`),
					helper.TestMatchResourceAttr("smallstep_collection_instance.thing1", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestMatchResourceAttr("smallstep_collection_instance.thing1", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
			},
			{
				Config: updated,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_collection_instance.thing1", "collection_slug", slug),
					helper.TestCheckResourceAttr("smallstep_collection_instance.thing1", "data", `{"name":"thing2"}`),
					helper.TestMatchResourceAttr("smallstep_collection_instance.thing1", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestMatchResourceAttr("smallstep_collection_instance.thing1", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
				),
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_collection_instance.thing1", plancheck.ResourceActionUpdate),
					},
				},
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
