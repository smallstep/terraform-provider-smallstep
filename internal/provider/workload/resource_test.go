package workload

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

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

func TestAccWorkloadResource(t *testing.T) {
	credential := utils.NewCredential(t)
	credentialID := *credential.Id
	workloadName := "tfprovider-" + utils.Slug(t)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(`
resource "smallstep_workload" "test" {
  name            = %[1]q
  credentials = [{
    credential_id = %[2]q
    probes = [{
      target   = "app.local:443"
      protocol = "TLS"
    }]
  }]
  workload_type   = "nginx"
  hooks = {
    sign = {
      after = ["systemctl reload nginx"]
      shell = "/bin/bash"
    }
  }
}
`, workloadName, credentialID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_workload.test", "id", utils.UUIDRegexp),
					helper.TestCheckResourceAttr("smallstep_workload.test", "name", workloadName),
					helper.TestCheckResourceAttr("smallstep_workload.test", "workload_type", "nginx"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "credentials.0.credential_id", credentialID),
					helper.TestCheckResourceAttr("smallstep_workload.test", "credentials.0.probes.0.target", "app.local:443"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "credentials.0.probes.0.protocol", "TLS"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "hooks.sign.after.0", "systemctl reload nginx"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "hooks.sign.shell", "/bin/bash"),
				),
			},
			{
				Config: fmt.Sprintf(`
resource "smallstep_workload" "test" {
  name            = %[1]q
  credentials = [{
    credential_id = %[2]q
    probes = [{
      target   = "app.local:8443"
      protocol = "TLS"
      server_name = "app.local"
    }]
  }]
  workload_type   = "systemd"
  hooks = {
    renew = {
      before = ["systemctl stop myservice"]
      after  = ["systemctl start myservice"]
      shell  = "/bin/bash"
    }
  }
  reload_info = {
    method    = "DBUS"
    unit_name = "nginx.service"
  }
}
`, workloadName, credentialID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_workload.test", "workload_type", "systemd"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "credentials.0.credential_id", credentialID),
					helper.TestCheckResourceAttr("smallstep_workload.test", "credentials.0.probes.0.target", "app.local:8443"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "credentials.0.probes.0.server_name", "app.local"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "hooks.renew.before.0", "systemctl stop myservice"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "hooks.renew.after.0", "systemctl start myservice"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "hooks.renew.shell", "/bin/bash"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "reload_info.method", "DBUS"),
					helper.TestCheckResourceAttr("smallstep_workload.test", "reload_info.unit_name", "nginx.service"),
				),
			},
			{
				ResourceName:      "smallstep_workload.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
