package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/stretchr/testify/require"
)

func TestAccManagedConfigurationResource(t *testing.T) {
	attestorRoot, _ := utils.CACerts(t)
	slug := utils.Slug(t)
	hostID := uuid.New().String()
	config := fmt.Sprintf(`
resource "smallstep_collection" "tpms" {
	slug = %q
}

resource "smallstep_collection_instance" "server1" {
	id = "urn:ek:sha256:RAzbOveN1Y45fYubuTxu5jOXWtOK1HbfZ7yHjBuWlyE="
	data = "{}"
	collection_slug = smallstep_collection.tpms.slug
	depends_on = [smallstep_collection.tpms]
}

resource "smallstep_attestation_authority" "aa" {
	name = "tfprovider%s"
	catalog = smallstep_collection.tpms.slug
	attestor_roots = %q
	depends_on = [smallstep_collection.tpms]
}

resource "smallstep_authority" "agents" {
	subdomain = %q
	name = "Agents Authority"
	type = "devops"
	admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_provisioner" "agents" {
	authority_id = smallstep_authority.agents.id
	name = "Agents"
	type = "ACME_ATTESTATION"
	acme_attestation = {
		attestation_formats = ["tpm"]
		attestation_roots = [smallstep_attestation_authority.aa.root]
	}
}

resource "smallstep_provisioner_webhook" "devices" {
	authority_id = smallstep_authority.agents.id
	provisioner_id = smallstep_provisioner.agents.id
	name = "devices"
	kind = "ENRICHING"
	cert_type = "X509"
	server_type = "HOSTED_ATTESTATION"
	collection_slug = smallstep_collection.tpms.slug
	depends_on = [smallstep_collection.tpms]
}

resource "smallstep_agent_configuration" "agent1" {
	authority_id = smallstep_authority.agents.id
	provisioner_name = smallstep_provisioner.agents.name
	name = "Agent1"
	attestation_slug = smallstep_attestation_authority.aa.slug
	depends_on = [smallstep_provisioner.agents]
}

resource "smallstep_endpoint_configuration" "ep1" {
	name = "My DB"
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

resource "smallstep_endpoint_configuration" "ep2" {
	name = "SSH"
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


resource "smallstep_managed_configuration" "mc" {
	agent_configuration_id = smallstep_agent_configuration.agent1.id
	host_id = %q
	name = "Foo MC"
	managed_endpoints = [
		{
			endpoint_configuration_id = smallstep_endpoint_configuration.ep1.id
			x509_certificate_data = {
				common_name = "db"
				sans = [
					"db",
					"db.default",
					"db.default.svc",
					"db.defaulst.svc.cluster.local",
				]
			}
		},
	]
}

resource "smallstep_managed_configuration" "mc2" {
	agent_configuration_id = smallstep_agent_configuration.agent1.id
	name = "Multiple Endpoints"
	managed_endpoints = [
		{
			endpoint_configuration_id = smallstep_endpoint_configuration.ep1.id
			x509_certificate_data = {
				common_name = "db"
				sans = ["db.internal"]
			}
		},
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
`, slug, slug, attestorRoot, slug, hostID)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					// agent
					resource.TestMatchResourceAttr("smallstep_agent_configuration.agent1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestMatchResourceAttr("smallstep_agent_configuration.agent1", "authority_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_agent_configuration.agent1", "provisioner_name", "Agents"),

					// x509 endpoint
					resource.TestMatchResourceAttr("smallstep_endpoint_configuration.ep1", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "name", "My DB"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "kind", "WORKLOAD"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.type", "X509"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.duration", "168h"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.crt_file", "db.crt"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.key_file", "db.key"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.root_file", "ca.crt"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.uid", "1001"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.gid", "999"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "certificate_info.mode", "256"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.shell", "/bin/sh"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.before.#", "3"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.after.#", "3"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.renew.on_error.#", "3"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.shell", "/bin/bash"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.before.#", "3"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.after.#", "3"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "hooks.sign.on_error.#", "3"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.type", "ECDSA_P256"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.format", "DER"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "key_info.pub_file", "file.csr"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "reload_info.method", "SIGNAL"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "reload_info.pid_file", "db.pid"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep1", "reload_info.signal", "1"),
					// ssh endpoint
					resource.TestMatchResourceAttr("smallstep_endpoint_configuration.ep2", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep2", "name", "SSH"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep2", "kind", "PEOPLE"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep2", "certificate_info.type", "SSH_USER"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep2", "key_info.type", "DEFAULT"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep2", "key_info.format", "DEFAULT"),
					resource.TestCheckResourceAttr("smallstep_endpoint_configuration.ep2", "reload_info.method", "AUTOMATIC"),
					// managed configuration
					resource.TestMatchResourceAttr("smallstep_managed_configuration.mc", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_managed_configuration.mc", "host_id", hostID),
					resource.TestCheckResourceAttr("smallstep_managed_configuration.mc", "name", "Foo MC"),
					resource.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.#", "1"),
					resource.TestMatchResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.endpoint_configuration_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.common_name", "db"),
					resource.TestCheckResourceAttr("smallstep_managed_configuration.mc", "managed_endpoints.0.x509_certificate_data.sans.#", "4"),
					// managed configuration 2
					resource.TestMatchResourceAttr("smallstep_managed_configuration.mc2", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestMatchResourceAttr("smallstep_managed_configuration.mc2", "host_id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_managed_configuration.mc2", "name", "Multiple Endpoints"),
					resource.TestCheckResourceAttr("smallstep_managed_configuration.mc2", "managed_endpoints.#", "2"),
				),
			},
			{
				ResourceName:      "smallstep_agent_configuration.agent1",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:            "smallstep_endpoint_configuration.ep1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_info.duration"},
			},
			{
				ResourceName:      "smallstep_endpoint_configuration.ep2",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "smallstep_managed_configuration.mc",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      "smallstep_managed_configuration.mc2",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
