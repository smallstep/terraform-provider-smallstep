package endpoint_configuration

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/authority"
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
	},
	DataSourceFactories: []func() datasource.DataSource{
		NewDataSource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccAgentConfigurationResource(t *testing.T) {
	root, _ := utils.CACerts(t)
	slug := utils.Slug(t)
	config1 := fmt.Sprintf(`
resource "smallstep_authority" "agents" {
	subdomain = %q
	name = "tfprovider-agents-authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "agents" {
	authority_id = smallstep_authority.agents.id
	name = "Agents"
	type = "X5C"
	x5c = {
		roots = [%q]
	}
}

resource "smallstep_endpoint_configuration" "ep1" {
	name = "tfprovider My DB"
	kind = "WORKLOAD"

	# this would generally be a different authority
	authority_id = smallstep_authority.agents.id

	# this would generally be an x5c provisioner
	provisioner_name = smallstep_provisioner.agents.name

	certificate_info = {
		type = "X509"
		duration = "168h"
		crt_file = "db.crt"
		key_file = "db.key"
		root_file = "ca.crt"
		uid = 1001
		gid = 999
		mode = 256
	}

	hooks = {
		renew = {
			shell = "/bin/sh"
			before = [
				"echo renewing 1",
				"echo renewing 2",
				"echo renewing 3",
			]
			after = [
				"echo renewed 1",
				"echo renewew 2",
				"echo renewed 3",
			]
			on_error = [
				"echo failed renew 1",
				"echo failed renew 2",
				"echo failed renew 3",
			]
		}
		sign = {
			shell = "/bin/bash"
			before = [
				"echo signing 1",
				"echo signing 2",
				"echo signing 3",
			]
			after = [
				"echo signed 1",
				"echo signed 2",
				"echo signed 3",
			]
			on_error = [
				"echo failed sign 1",
				"echo failed sign 2",
				"echo failed sign 3",
			]
		}
	}

	key_info = {
		format = "DER"
		type = "ECDSA_P256"
		pub_file = "file.csr"
	}


	reload_info = {
		method = "SIGNAL"
		pid_file = "db.pid"
		signal = 1
	}
}
`, slug, root)

	slug2 := utils.Slug(t)
	config2 := fmt.Sprintf(`
resource "smallstep_authority" "agents" {
	subdomain = %q
	name = "tfprovider-agents-authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "agents" {
	authority_id = smallstep_authority.agents.id
	name = "Agents"
	type = "X5C"
	x5c = {
		roots = [%q]
	}
}

resource "smallstep_endpoint_configuration" "ep1" {
	name = "tfprovider SSH"
	kind = "PEOPLE"
	authority_id = smallstep_authority.agents.id
	provisioner_name = smallstep_provisioner.agents.name
	certificate_info = {
		type = "SSH_USER"
	}

	key_info = {
		type = "DEFAULT"
		format = "DEFAULT"
	}

	reload_info = {
		method = "AUTOMATIC"
	}
	hooks = {
		sign = {}
	}
}
`, slug2, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config1,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_endpoint_configuration.ep1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "name", "tfprovider My DB"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "kind", "WORKLOAD"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.type", "X509"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.duration", "168h"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.crt_file", "db.crt"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.key_file", "db.key"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.root_file", "ca.crt"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.uid", "1001"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.gid", "999"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.mode", "256"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.shell", "/bin/sh"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.before.#", "3"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.after.#", "3"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.on_error.#", "3"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.shell", "/bin/bash"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.before.#", "3"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.after.#", "3"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.on_error.#", "3"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.type", "ECDSA_P256"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.format", "DER"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.pub_file", "file.csr"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "reload_info.method", "SIGNAL"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "reload_info.pid_file", "db.pid"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "reload_info.signal", "1"),
				),
			},
			{
				ResourceName:            "smallstep_endpoint_configuration.ep1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_info.duration"},
			},
			{
				Config: config2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_endpoint_configuration.ep1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "name", "tfprovider SSH"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "kind", "PEOPLE"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.type", "SSH_USER"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.type", "DEFAULT"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.format", "DEFAULT"),
					helper.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "reload_info.method", "AUTOMATIC"),
				),
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_endpoint_configuration.ep1", plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:      "smallstep_endpoint_configuration.ep1",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
