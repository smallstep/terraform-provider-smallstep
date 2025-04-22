package device

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
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

const minConfig = `
resource "smallstep_device" "laptop1" {
	permanent_identifier = %q
}`

const maxConfig = `
resource "smallstep_device" "laptop1" {
	permanent_identifier = %q
	display_id = "9 9"
	display_name = "Employee Laptop"
	serial = "678"
	tags = ["ubuntu"]
	metadata = {
		k1 = "v1"
	}
	os = "Windows"
	ownership = "company"
	user = {
		email = "user@example.com"
	}
}`

const emptyConfig = `
resource "smallstep_device" "laptop1" {
	permanent_identifier = %q
	display_id = ""
	display_name = ""
	serial = ""
	tags = []
	metadata = {}
}
`

func TestAccDeviceResource(t *testing.T) {
	permanentID := uuid.NewString()

	// min -> max
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(minConfig, permanentID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID),
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "host_id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "display_id"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "display_name"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "serial"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "user"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "os"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "ownership"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "metadata"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "tags"),
				),
			},
			{
				ResourceName:      "smallstep_device.laptop1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(maxConfig, permanentID),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_id", "9 9"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_name", "Employee Laptop"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "serial", "678"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "user.email", "user@example.com"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "os", "Windows"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "ownership", "company"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "tags.0", "ubuntu"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "metadata.k1", "v1"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
				),
			},
		},
	})

	// min -> empty
	permanentID1 := uuid.NewString()
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(minConfig, permanentID1),
			},
			{
				Config: fmt.Sprintf(emptyConfig, permanentID1),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID1),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_id", ""),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_name", ""),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "serial", ""),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "user"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "os"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "ownership"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "metadata.%", "0"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "tags.%", "0"),
				),
			},
		},
	})

	permanentID2 := uuid.NewString()

	// max -> min
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(maxConfig, permanentID2),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID2),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_id", "9 9"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_name", "Employee Laptop"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "serial", "678"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "user.email", "user@example.com"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "os", "Windows"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "ownership", "company"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "tags.0", "ubuntu"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "metadata.k1", "v1"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
				),
			},
			{
				ResourceName:      "smallstep_device.laptop1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(minConfig, permanentID2),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID2),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "display_id"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "display_name"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "serial"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "os"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "ownership"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "user.email"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "metadata"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "tags"),
				),
			},
		},
	})

	permanentID3 := uuid.NewString()

	// max -> empty
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(maxConfig, permanentID3),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID3),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_id", "9 9"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_name", "Employee Laptop"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "serial", "678"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "user.email", "user@example.com"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "os", "Windows"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "ownership", "company"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "tags.0", "ubuntu"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "metadata.k1", "v1"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
				),
			},
			{
				ResourceName:      "smallstep_device.laptop1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(emptyConfig, permanentID3),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID3),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_id", ""),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_name", ""),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "serial", ""),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "user"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "os"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "ownership"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "metadata.%", "0"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "tags.%", "0"),
				),
			},
		},
	})

	permanentID4 := uuid.NewString()

	// empty -> min
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(emptyConfig, permanentID4),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID4),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_id", ""),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_name", ""),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "serial", ""),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "user"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "os"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "ownership"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "metadata.%", "0"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "tags.%", "0"),
				),
			},
			{
				ResourceName:            "smallstep_device.laptop1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"display_id", "display_name", "serial"},
			},
			{
				Config: fmt.Sprintf(minConfig, permanentID4),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID4),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "display_id"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "display_name"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "serial"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "user"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "os"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "ownership"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "metadata"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "tags"),
				),
			},
		},
	})

	// empty -> max
	permanentID5 := uuid.NewString()

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fmt.Sprintf(emptyConfig, permanentID5),
			},
			{
				Config: fmt.Sprintf(maxConfig, permanentID5),
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", utils.UUID),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID5),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_id", "9 9"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "display_name", "Employee Laptop"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "serial", "678"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "user.email", "user@example.com"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "os", "Windows"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "ownership", "company"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "tags.0", "ubuntu"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "metadata.k1", "v1"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "connected", "false"),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "high_assurance", "false"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "enrolled_at"),
					helper.TestCheckNoResourceAttr("smallstep_device.laptop1", "last_seen"),
				),
			},
		},
	})
}
