package workload

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccWorkloadDataSource(t *testing.T) {
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

data "smallstep_workload" "test" {
  id = smallstep_workload.test.id
}
`, workloadName, credentialID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttrSet("data.smallstep_workload.test", "id"),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "name", workloadName),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "workload_type", "nginx"),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "credentials.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "credentials.0.credential_id", credentialID),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "credentials.0.probes.0.target", "app.local:443"),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "credentials.0.probes.0.protocol", "TLS"),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "hooks.sign.after.0", "systemctl reload nginx"),
					helper.TestCheckResourceAttr("data.smallstep_workload.test", "hooks.sign.shell", "/bin/bash"),
				),
			},
		},
	})
}
