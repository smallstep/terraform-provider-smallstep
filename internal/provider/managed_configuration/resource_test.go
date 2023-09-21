package managed_configuration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/agent_configuration"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/authority"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/endpoint_configuration"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/provisioner"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
		authority.NewResource,
		provisioner.NewResource,
		agent_configuration.NewResource,
		endpoint_configuration.NewResource,
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccManagedConfigurationResource(t *testing.T) {
	root, _ := utils.CACerts(t)
	slug := utils.Slug(t)
	hostID := uuid.New().String()
	config := fmt.Sprintf(`
resource "smallstep_authority" "authority" {
	subdomain = %q
	name = "tfprovider-managed-workloads-authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "provisioner" {
	authority_id = smallstep_authority.authority.id
	name = "Managed Workloads Provisioner"
	type = "X5C"
	x5c = {
		roots = [%q]
	}
}

resource "smallstep_agent_configuration" "agent1" {
	authority_id = smallstep_authority.authority.id
	provisioner_name = smallstep_provisioner.provisioner.name
	name = "tfprovider Agent1"
	attestation_slug = "attestationslug"
	depends_on = [smallstep_provisioner.provisioner]
}

resource "smallstep_endpoint_configuration" "ep1" {
	name = "tfprovider My DB"
	kind = "WORKLOAD"
	authority_id = smallstep_authority.authority.id
	provisioner_name = smallstep_provisioner.provisioner.name

	certificate_info = {
		type = "X509"
	}

	key_info = {
		format = "DEFAULT"
		type = "DEFAULT"
	}
}

resource "smallstep_managed_configuration" "mc" {
	agent_configuration_id = smallstep_agent_configuration.agent1.id
	name = "tfprovider Multiple Endpoints"
	host_id = %q
	managed_endpoints = [
		{
			endpoint_configuration_id = smallstep_endpoint_configuration.ep1.id
			x509_certificate_data = {
				common_name = "db"
				sans = ["db", "db.internal"]
			}
		},
	]
}
`, slug, root, hostID)

	hostID2 := uuid.New().String()
	config2 := fmt.Sprintf(`
resource "smallstep_authority" "authority" {
	subdomain = %q
	name = "tfprovider-managed-workloads-authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "provisioner" {
	authority_id = smallstep_authority.authority.id
	name = "Managed Workloads Provisioner"
	type = "X5C"
	x5c = {
		roots = [%q]
	}
}

resource "smallstep_agent_configuration" "agent2" {
	authority_id = smallstep_authority.authority.id
	provisioner_name = smallstep_provisioner.provisioner.name
	name = "tfprovider Agent1"
	attestation_slug = "attestationslug"
	depends_on = [smallstep_provisioner.provisioner]
}


resource "smallstep_endpoint_configuration" "ep2" {
	name = "tfprovider SSH"
	kind = "PEOPLE"
	authority_id = smallstep_authority.authority.id
	provisioner_name = smallstep_provisioner.provisioner.name

	certificate_info = {
		type = "SSH_USER"
	}

	key_info = {
		type = "DEFAULT"
		format = "DEFAULT"
	}
}

resource "smallstep_managed_configuration" "mc" {
	agent_configuration_id = smallstep_agent_configuration.agent2.id
	name = "tfprovider Updated"
	host_id = %q
	managed_endpoints = [
		{
			endpoint_configuration_id = smallstep_endpoint_configuration.ep2.id
			ssh_certificate_data = {
				key_id = "abc"
				principals = [
					"ops",
					"eng",
					"sec",
				]
			}
		},
	]
}
`, slug, root, hostID2)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config,
				Check: helper.ComposeAggregateTestCheckFunc(
					// managed configuration
					helper.TestMatchResourceAttr("smallstep_managed_configuration.mc", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "host_id", hostID),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "name", "tfprovider Multiple Endpoints"),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.#", "1"),
					helper.TestMatchResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.endpoint_configuration_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.common_name", "db"),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.sans.#", "2"),
				),
			},
			{
				ResourceName:      "smallstep_managed_configuration.mc",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: config2,
				Check: helper.ComposeAggregateTestCheckFunc(
					// managed configuration
					helper.TestMatchResourceAttr("smallstep_managed_configuration.mc", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "host_id", hostID2),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "name", "tfprovider Updated"),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.#", "1"),
					helper.TestMatchResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.endpoint_configuration_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.ssh_certificate_data.key_id", "abc"),
					helper.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.ssh_certificate_data.principals.#", "3"),
				),
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_managed_configuration.mc", plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
