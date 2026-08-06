package managed_radius

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccManagedRadiusDataSource(t *testing.T) {
	radius := utils.NewManagedRADIUS(t)

	config := fmt.Sprintf(`
data "smallstep_managed_radius" "my_rad" {
	id = %q
}`, *radius.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "id", *radius.Id),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "name", radius.Name),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "client_ca", radius.ClientCA),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "nas_ips.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "nas_ips.0", radius.NasIPs[0]),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "reply_attributes.#", "2"),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "reply_attributes.0.name", "Tunnel-Type"),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "reply_attributes.0.value", "13"),
					helper.TestCheckNoResourceAttr("data.smallstep_managed_radius.my_rad", "reply_attributes.0.value_from_certificate"),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "reply_attributes.1.name", "Tunnel-Private-Group-ID"),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "reply_attributes.1.value_from_certificate", "2.5.4.11"),
					helper.TestCheckNoResourceAttr("data.smallstep_managed_radius.my_rad", "reply_attributes.1.value"),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "server_ca", *radius.ServerCA),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "server_hostname", *radius.ServerHostname),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "server_ip", *radius.ServerIP),
					helper.TestCheckResourceAttr("data.smallstep_managed_radius.my_rad", "server_port", *radius.ServerPort),
				),
			},
		},
	})
}
