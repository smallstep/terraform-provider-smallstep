package managed_radius

import (
	"fmt"
	"regexp"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccManagedRadius(t *testing.T) {
	ip := utils.IP(t)
	ip2 := utils.IP(t)
	ca, _ := utils.CACerts(t)
	ca2, _ := utils.CACerts(t)
	name := "My Managed Radius"
	name2 := "My Managed RADIUS"

	const config = `
resource "smallstep_managed_radius" "my_rad" {
	name = %q
	nas_ips = [%q]
	client_ca = %q
	reply_attributes = [{
		name = "Tunnel-Type"
		value = "13"
	}]
}`
	const config2 = `
resource "smallstep_managed_radius" "my_rad" {
	name = %q
	nas_ips = [%q]
	client_ca = %q
	reply_attributes = [{
		name = "Tunnel-Private-Group-ID"
		value_from_extension = "2.5.4.11"
	}]
}`
	cfg := fmt.Sprintf(config, name, ip, ca)
	cfg2 := fmt.Sprintf(config2, name2, ip2, ca2)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: cfg,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_managed_radius.my_rad", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "name", name),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "client_ca", ca),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "nas_ips.#", "1"),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "nas_ips.0", ip),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "1"),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.0.name", "Tunnel-Type"),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.0.value", "13"),
					helper.TestCheckNoResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.0.value_from_extension"),
					helper.TestMatchResourceAttr("smallstep_managed_radius.my_rad", "server_ca", utils.CARegexp),
					helper.TestMatchResourceAttr("smallstep_managed_radius.my_rad", "server_hostname", regexp.MustCompile(`^radius\.\w+`)),
					helper.TestMatchResourceAttr("smallstep_managed_radius.my_rad", "server_ip", utils.IPv4Regexp),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "server_port", "1812"),
				),
			},
			{
				Config: cfg2,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_managed_radius.my_rad", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_managed_radius.my_rad", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "name", name2),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "client_ca", ca2),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "nas_ips.#", "1"),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "nas_ips.0", ip2),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "1"),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.0.name", "Tunnel-Private-Group-ID"),
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.0.value_from_extension", "2.5.4.11"),
					helper.TestCheckNoResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.0.value"),
				),
			},
			{
				ResourceName:      "smallstep_managed_radius.my_rad",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	const empty = `
resource "smallstep_managed_radius" "my_rad" {
	name = %q
	nas_ips = [%q]
	client_ca = %q
	reply_attributes = []
}`
	emptyConfig := fmt.Sprintf(empty, name, ip, ca)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "0"),
				),
			},
			{
				ResourceName:      "smallstep_managed_radius.my_rad",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: cfg,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_managed_radius.my_rad", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "1"),
				),
			},
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "0"),
				),
			},
		},
	})

	const min = `
resource "smallstep_managed_radius" "my_rad" {
	name = %q
	nas_ips = [%q]
	client_ca = %q
}`
	minConfig := fmt.Sprintf(min, name, ip, ca)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "0"),
				),
			},
			{
				ResourceName:      "smallstep_managed_radius.my_rad",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: cfg,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_managed_radius.my_rad", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "1"),
				),
			},
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "0"),
				),
			},
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "0"),
				),
			},
			{
				Config: minConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_managed_radius.my_rad", "reply_attributes.#", "0"),
				),
			},
		},
	})
}
