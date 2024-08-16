package device_collection_account

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
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
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccAccountResource(t *testing.T) {
	t.Parallel()

	dc := utils.NewTPMDeviceCollection(t)
	account, _ := utils.NewAccount(t)
	authority := utils.NewAuthority(t)
	slug := utils.Slug(t)

	config := fmt.Sprintf(`
resource "smallstep_device_collection_account" "test" {
	slug = %q
	device_collection_slug = %q
	account_id = %q
	authority_id = %q
	display_name = "Tester"
	certificate_info = {
		type = "X509"
	}
	key_info = {
		type = "ECDSA_P256"
		protection = "HARDWARE"
		format = "DEFAULT"
	}
	certificate_data = {
		common_name = {
			device_metadata = "email"
		}
		sans = {
			device_metadata = ["email"]
		}
	}
}`, slug, dc.Slug, *account.Id, authority.Id)

	config2 := fmt.Sprintf(`
resource "smallstep_device_collection_account" "test" {
	slug = %q
	device_collection_slug = %q
	account_id = %q
	authority_id = %q
	display_name = "Tester 2"
	certificate_info = {
		type = "X509"
	}
	key_info = {
		type = "ECDSA_P256"
		protection = "HARDWARE"
		format = "DEFAULT"
	}
	certificate_data = {
		common_name = {
			device_metadata = "email"
		}
		sans = {
			device_metadata = ["email"]
		}
	}
}`, slug, dc.Slug, *account.Id, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_device_collection_account.test", plancheck.ResourceActionCreate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection_account.test", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection_account.test", "display_name", "Tester"),
				),
			},
			{
				Config: config2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_device_collection_account.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection_account.test", "display_name", "Tester 2"),
				),
			},
			{
				ResourceName:                         "smallstep_device_collection_account.test",
				ImportState:                          true,
				ImportStateId:                        fmt.Sprintf("%s/%s", dc.Slug, slug),
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "slug",
			},
		},
	})
}
