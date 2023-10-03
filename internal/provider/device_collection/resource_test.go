package device_collection

import (
	"fmt"
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

func TestAccDeviceCollectionResource(t *testing.T) {
	slug := utils.Slug(t)
	awsRequired := fmt.Sprintf(`
resource "smallstep_device_collection" "aws_required_only" {
	slug = %q
	display_name = "EC2 West"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
	}
}`, slug)

	updated := fmt.Sprintf(`
resource "smallstep_device_collection" "aws_required_only" {
	slug = %q
	display_name = "EC2 East"
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
				Config: awsRequired,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "display_name", "EC2 West"),
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "aws_vm.accounts.0", "0123456789"),
				),
			},
			{
				Config: updated,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "display_name", "EC2 East"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	awsOptionalEmpty := fmt.Sprintf(`
resource "smallstep_device_collection" "aws_optional_empty" {
	slug = %q
	display_name = "EC2 West"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
		disable_custom_sans = false
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsOptionalEmpty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_optional_empty", "aws_vm.disable_custom_sans", "false"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	awsOptionalNonempty := fmt.Sprintf(`
resource "smallstep_device_collection" "aws_optional_nonempty" {
	slug = %q
	display_name = "EC2 West"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
		disable_custom_sans = true
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsOptionalNonempty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_optional_nonempty", "aws_vm.disable_custom_sans", "true"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	gcpRequired := fmt.Sprintf(`
resource "smallstep_device_collection" "gcp_required_only" {
	slug = %q
	display_name = "GCE"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "gcp-vm"
	gcp_vm = {
		service_accounts = ["0123456789"]
		project_ids = ["prod-1"]
	}
}`, slug)

	updatedGCPRequired := fmt.Sprintf(`
resource "smallstep_device_collection" "gcp_required_only" {
	slug = %q
	display_name = "Google Compute Engine"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "gcp-vm"
	gcp_vm = {
		service_accounts = ["0123456789"]
		project_ids = ["prod-1"]
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpRequired,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "display_name", "GCE"),
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "gcp_vm.service_accounts.0", "0123456789"),
				),
			},
			{
				Config: updatedGCPRequired,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "display_name", "Google Compute Engine"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	gcpOptionalEmpty := fmt.Sprintf(`
resource "smallstep_device_collection" "gcp_optional_empty" {
	slug = %q
	display_name = "GCE"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "gcp-vm"
	gcp_vm = {
		service_accounts = ["0123456789"]
		disable_custom_sans = false
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpOptionalEmpty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_optional_empty", "gcp_vm.disable_custom_sans", "false"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	gcpOptionalNonempty := fmt.Sprintf(`
resource "smallstep_device_collection" "gcp_optional_nonempty" {
	slug = %q
	display_name = "GCE"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "gcp-vm"
	gcp_vm = {
		project_ids = ["prod-123"]
		disable_custom_sans = true
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpOptionalNonempty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_optional_nonempty", "gcp_vm.disable_custom_sans", "true"),
				),
			},
		},
	})
}
