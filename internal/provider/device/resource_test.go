package device

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
	},
	/*
		DataSourceFactories: []func() datasource.DataSource{
			NewDataSource,
		},
	*/
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccAuthorityResource(t *testing.T) {
	t.Parallel()

	permanentID := uuid.NewString()

	emptyConfig := fmt.Sprintf(`
resource "smallstep_device" "laptop1" {
	permanent_identifier = %q
}`, permanentID)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: emptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_device.laptop1", "permanent_identifier", permanentID),
				),
			},
			{
				ResourceName:      "smallstep_device.laptop1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})

	permanentID2 := uuid.NewString()
	fullConfig := fmt.Sprintf(`
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
}`, permanentID2)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: fullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
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
		},
	})

	permanentID3 := uuid.NewString()
	config1 := fmt.Sprintf(`
resource "smallstep_device" "laptop1" {
	permanent_identifier = %q
}`, permanentID3)
	config2 := fmt.Sprintf(`
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
}`, permanentID3)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config1,
			},
			{
				Config: config2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_device.laptop1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
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
		},
	})
}
