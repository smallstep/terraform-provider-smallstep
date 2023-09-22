package device_collection

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccCollectionResource(t *testing.T) {
	t.Parallel()

	slug := utils.Slug(t)
	config := fmt.Sprintf(`
resource "smallstep_device_collection" "ec2_west" {
	slug = %q
	display_name = "EC2 West"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.ec2_west", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.ec2_west", "display_name", "EC2 West"),
					helper.TestCheckResourceAttr("smallstep_device_collection.ec2_west", "schema_uri", "https://schema.infra.smallstep.com/managed-workloads/device/aws-vm"),
					helper.TestCheckResourceAttr("smallstep_device_collection.ec2_west", "instance_count", "0"),
					helper.TestMatchResourceAttr("smallstep_device_collection.ec2_west", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestMatchResourceAttr("smallstep_device_collection.ec2_west", "updated_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_device_collection.ec2_west", "aws_vm.accounts.0", "0123456789"),
				),
			},
		},
	})
}
