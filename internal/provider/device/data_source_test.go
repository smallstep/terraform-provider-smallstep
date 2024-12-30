package device

import (
	"fmt"
	"testing"

	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccProvisionerDataSource(t *testing.T) {
	t.Parallel()
	device := utils.NewDevice(t)

	config := fmt.Sprintf(`
data "smallstep_device" "test" {
	id = %q
}`, device.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("data.smallstep_device.test", "id", device.Id),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "permanent_identifier", device.PermanentIdentifier),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "display_id", utils.Deref(device.DisplayId)),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "display_name", utils.Deref(device.DisplayName)),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "serial", utils.Deref(device.Serial)),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "os", string(utils.Deref(device.Os))),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "ownership", string(utils.Deref(device.Ownership))),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "user.email", device.User.Email),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "tags.#", "1"),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "metadata.%", "1"),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "high_assurance", "false"),
					helper.TestCheckResourceAttr("data.smallstep_device.test", "connected", "false"),
					helper.TestCheckNoResourceAttr("data.smallstep_device.test", "enrolled_at"),
					helper.TestCheckNoResourceAttr("data.smallstep_device.test", "last_seen"),
				),
			},
		},
	})
}
