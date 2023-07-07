package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccManagedWorkloadDataSource(t *testing.T) {
	authority := utils.NewAuthority(t)
	provisioner, _ := utils.NewOIDCProvisioner(t, authority.Id)
	collection := utils.NewCollection(t)
	attest := utils.FixAttestationAuthority(t, collection.Slug)
	ac := utils.NewAgentConfiguration(t, authority.Id, provisioner.Name, *attest.Slug)
	ec := utils.NewEndpointConfiguration(t, authority.Id, provisioner.Name)
	mc := utils.NewManagedConfiguration(t, *ac.Id, *ec.Id)

	agentConfig := fmt.Sprintf(`
data "smallstep_agent_configuration" "agent" {
	id = %q
}
`, *ac.Id)

	managedConfig := fmt.Sprintf(`
data "smallstep_managed_configuration" "mc" {
	id = %q
}
`, *mc.Id)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: managedConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "id", *mc.Id),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "agent_configuration_id", *ac.Id),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "host_id", utils.Deref(mc.HostID)),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "name", mc.Name),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.#", "1"),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.endpoint_configuration_id", *ec.Id),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.id", *mc.ManagedEndpoints[0].Id),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.common_name", mc.ManagedEndpoints[0].X509CertificateData.CommonName),
					resource.TestCheckResourceAttr("data.smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.sans.#", "1"),
				),
			},
		},
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: agentConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "id", *ac.Id),
					resource.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "authority_id", authority.Id),
					resource.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "name", ac.Name),
					resource.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "provisioner_name", ac.Provisioner),
					resource.TestCheckResourceAttr("data.smallstep_agent_configuration.agent", "attestation_slug", *ac.AttestationSlug),
				),
			},
		},
	})

	endpointConfig := fmt.Sprintf(`
data "smallstep_endpoint_configuration" "ep" {
	id = %q
}
`, *ec.Id)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: endpointConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("data.smallstep_endpoint_configuration.ep", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "authority_id", authority.Id),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "name", ec.Name),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "provisioner_name", ec.Provisioner),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.type", string(ec.CertificateInfo.Type)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.duration", utils.Deref(ec.CertificateInfo.Duration)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.crt_file", utils.Deref(ec.CertificateInfo.CrtFile)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.root_file", utils.Deref(ec.CertificateInfo.RootFile)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.key_file", utils.Deref(ec.CertificateInfo.KeyFile)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.uid", strconv.Itoa(utils.Deref(ec.CertificateInfo.Uid))),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.gid", strconv.Itoa(utils.Deref(ec.CertificateInfo.Gid))),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.mode", strconv.Itoa(utils.Deref(ec.CertificateInfo.Mode))),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.shell", utils.Deref(ec.Hooks.Sign.Shell)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.before.#", "1"),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.before.0", (*ec.Hooks.Sign.Before)[0]),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.after.#", "1"),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.after.0", (*ec.Hooks.Sign.After)[0]),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.on_error.#", "1"),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.on_error.0", (*ec.Hooks.Sign.OnError)[0]),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.shell", utils.Deref(ec.Hooks.Renew.Shell)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.before.#", "1"),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.before.0", (*ec.Hooks.Renew.Before)[0]),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.after.#", "1"),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.after.0", (*ec.Hooks.Renew.After)[0]),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.on_error.#", "1"),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.on_error.0", (*ec.Hooks.Renew.OnError)[0]),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "key_info.type", string(utils.Deref(ec.KeyInfo.Type))),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "key_info.format", string(utils.Deref(ec.KeyInfo.Format))),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "key_info.pub_file", utils.Deref(ec.KeyInfo.PubFile)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "reload_info.method", string(ec.ReloadInfo.Method)),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "reload_info.signal", strconv.Itoa(utils.Deref(ec.ReloadInfo.Signal))),
					resource.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "reload_info.pid_file", utils.Deref(ec.ReloadInfo.PidFile)),
				),
			},
		},
	})
}
