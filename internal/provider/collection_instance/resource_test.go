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
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/device_collection"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		collection.NewResource,
		device_collection.NewResource,
		NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccDeviceInstance(t *testing.T) {
	slug := utils.Slug(t)
	config1 := fmt.Sprintf(`
resource "smallstep_device_collection" "ec2_east" {
	slug = %q
	display_name = "EC2 East"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
	}
}

resource "smallstep_collection_instance" "thing1" {
	depends_on = [smallstep_device_collection.ec2_east]
	collection_slug = smallstep_device_collection.ec2_east.slug
	id = "i-%s"
	data = "{\"name\":\"thing1\"}"
}`, slug, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config1,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_collection_instance.thing1", "id", "i-"+slug),
				),
			},
		},
	})
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
	id = %q
	data = "{\"name\":\"thing1\"}"
	depends_on = [smallstep_collection.things]
}`, slug, slug, slug)

	updated := fmt.Sprintf(`
resource "smallstep_collection" "things" {
	slug = %q
}

resource "smallstep_collection_instance" "thing1" {
	collection_slug = %q
	id = %q
	data = "{\"name\":\"thing2\"}"
	depends_on = [smallstep_collection.things]
}`, slug, slug, slug)

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
				ImportStateId: fmt.Sprintf("%s/%s", slug, slug),
				// This only succeeds because the data JSON is simple enough to
				// have the same serialization every time. Change to `false` to
				// test more complex JSON.
				ImportStateVerify: true,
			},
		},
	})
}
