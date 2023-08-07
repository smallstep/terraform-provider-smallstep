package endpoint_configuration

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccManagedWorkloadDataSource(t *testing.T) {
	authority := utils.NewAuthority(t)
	provisioner, _ := utils.NewOIDCProvisioner(t, authority.Id)
	ec := utils.NewEndpointConfiguration(t, authority.Id, provisioner.Name)

	endpointConfig := fmt.Sprintf(`
data "smallstep_endpoint_configuration" "ep" {
	id = %q
}
`, *ec.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: endpointConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("data.smallstep_endpoint_configuration.ep", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "authority_id", authority.Id),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "name", ec.Name),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "provisioner_name", ec.Provisioner),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.type", string(ec.CertificateInfo.Type)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.duration", utils.Deref(ec.CertificateInfo.Duration)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.crt_file", utils.Deref(ec.CertificateInfo.CrtFile)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.root_file", utils.Deref(ec.CertificateInfo.RootFile)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.key_file", utils.Deref(ec.CertificateInfo.KeyFile)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.uid", strconv.Itoa(utils.Deref(ec.CertificateInfo.Uid))),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.gid", strconv.Itoa(utils.Deref(ec.CertificateInfo.Gid))),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "certificate_info.mode", strconv.Itoa(utils.Deref(ec.CertificateInfo.Mode))),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.shell", utils.Deref(ec.Hooks.Sign.Shell)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.before.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.before.0", (*ec.Hooks.Sign.Before)[0]),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.after.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.after.0", (*ec.Hooks.Sign.After)[0]),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.on_error.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.sign.on_error.0", (*ec.Hooks.Sign.OnError)[0]),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.shell", utils.Deref(ec.Hooks.Renew.Shell)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.before.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.before.0", (*ec.Hooks.Renew.Before)[0]),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.after.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.after.0", (*ec.Hooks.Renew.After)[0]),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.on_error.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "hooks.renew.on_error.0", (*ec.Hooks.Renew.OnError)[0]),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "key_info.type", string(utils.Deref(ec.KeyInfo.Type))),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "key_info.format", string(utils.Deref(ec.KeyInfo.Format))),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "key_info.pub_file", utils.Deref(ec.KeyInfo.PubFile)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "reload_info.method", string(ec.ReloadInfo.Method)),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "reload_info.signal", strconv.Itoa(utils.Deref(ec.ReloadInfo.Signal))),
					helper.TestCheckResourceAttr("data.smallstep_endpoint_configuration.ep", "reload_info.pid_file", utils.Deref(ec.ReloadInfo.PidFile)),
				),
			},
		},
	})
}
