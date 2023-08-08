package managed_configuration

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccManagedWorkloadDataSource(t *testing.T) {
	authority := utils.NewAuthority(t)
	provisioner, _ := utils.NewOIDCProvisioner(t, authority.Id)
	ac := utils.NewAgentConfiguration(t, authority.Id, provisioner.Name, "attestslug")
	ec := utils.NewEndpointConfiguration(t, authority.Id, provisioner.Name)
	mc := utils.NewManagedConfiguration(t, *ac.Id, *ec.Id)

	managedConfig := fmt.Sprintf(`
data "smallstep_managed_configuration" "mc" {
	id = %q
}
`, *mc.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: managedConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "id", *mc.Id),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "agent_configuration_id", *ac.Id),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "host_id", utils.Deref(mc.HostID)),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "name", mc.Name),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.endpoint_configuration_id", *ec.Id),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.id", *mc.ManagedEndpoints[0].Id),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.common_name", mc.ManagedEndpoints[0].X509CertificateData.CommonName),
					helper.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.sans.#", "1"),
				),
			},
		},
	})
}
