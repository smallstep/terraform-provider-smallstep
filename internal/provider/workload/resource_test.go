package workload

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/device_collection"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
		device_collection.NewResource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccAgentConfigurationResource(t *testing.T) {
	dcSlug := utils.Slug(t)
	slug1 := utils.Slug(t)
	slug2 := utils.Slug(t)
	config1 := fmt.Sprintf(`
resource "smallstep_device_collection" "ec2_east" {
	slug = %q
	display_name = "EC2 East"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
	}
}

resource "smallstep_workload" "generic" {
	depends_on = [smallstep_device_collection.ec2_east]
	workload_type = "generic"
	admin_emails = ["andrew@smallstep.com"]
	device_collection_slug = resource.smallstep_device_collection.ec2_east.slug
	slug = %q
	display_name = "tfprovider generic"

	certificate_info = {
		type = "X509"
	}
	key_info = {
		format = "DER"
		type = "ECDSA_P256"
	}
	static_sans = ["host.internal"]
}

resource "smallstep_workload" "nginx" {
	depends_on = [smallstep_device_collection.ec2_east]
	workload_type = "nginx"
	admin_emails = ["andrew@smallstep.com"]
	device_collection_slug = resource.smallstep_device_collection.ec2_east.slug
	slug = %q
	display_name = "tfprovider Nginx"

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
`, dcSlug, slug1, slug2)

	config2 := fmt.Sprintf(`
resource "smallstep_device_collection" "ec2_east" {
	slug = %q
	display_name = "EC2 East"
	admin_emails = ["andrew@smallstep.com"]
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
	}
}

resource "smallstep_workload" "generic" {
	depends_on = [smallstep_device_collection.ec2_east]
	workload_type = "generic"
	admin_emails = ["andrew@smallstep.com"]
	device_collection_slug = resource.smallstep_device_collection.ec2_east.slug
	slug = %q
	display_name = "tfprovider generic"

	certificate_info = {
		type = "X509"
	}
	key_info = {
		format = "DER"
		type = "ECDSA_P256"
	}
	device_metadata_key_sans = ["internal_host"]
}

resource "smallstep_workload" "nginx" {
	depends_on = [smallstep_device_collection.ec2_east]
	workload_type = "nginx"
	admin_emails = ["andrew@smallstep.com"]
	device_collection_slug = resource.smallstep_device_collection.ec2_east.slug
	slug = %q
	display_name = "tfprovider Nginx"

	certificate_info = {
		type = "X509"
		duration = "167h"
		crt_file = "pg.crt"
		key_file = "pg.key"
		root_file = "pg_ca.crt"
		uid = 1002
		gid = 998
		mode = 7
	}

	hooks = {
		renew = {
			shell = "/bin/zsh"
			before = [
				"echo renewing 4",
			]
			after = [
				"echo renewed 4",
			]
			on_error = [
				"echo failed renew 4",
			]
		}
		sign = {
			shell = "/bin/fish"
			before = [
				"echo signing 4",
			]
			after = [
				"echo signed 4",
			]
			on_error = [
				"echo failed sign 4",
			]
		}
	}

	key_info = {
		format = "DEFAULT"
		type = "ECDSA_P384"
		pub_file = ""
	}

	reload_info = {
		method = "DBUS"
		unit_name = "postgres.service"
	}
}`, dcSlug, slug1, slug2)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config1,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "display_name", "tfprovider Nginx"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.type", "X509"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.duration", "168h"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.crt_file", "db.crt"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.key_file", "db.key"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.root_file", "ca.crt"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.uid", "1001"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.gid", "999"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.mode", "256"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.shell", "/bin/sh"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.before.#", "3"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.after.#", "3"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.on_error.#", "3"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.shell", "/bin/bash"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.before.#", "3"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.after.#", "3"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.on_error.#", "3"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "key_info.type", "ECDSA_P256"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "key_info.format", "DER"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "key_info.pub_file", "file.csr"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "reload_info.method", "SIGNAL"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "reload_info.pid_file", "db.pid"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "reload_info.signal", "1"),

					helper.TestCheckResourceAttr("smallstep_workload.generic", "static_sans.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.generic", "static_sans.0", "host.internal"),
					helper.TestCheckResourceAttr("smallstep_workload.generic", "device_metadata_key_sans.#", "0"),
				),
			},
		},
	})

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: config2,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "display_name", "tfprovider Nginx"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.type", "X509"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.duration", "167h"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.crt_file", "pg.crt"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.key_file", "pg.key"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.root_file", "pg_ca.crt"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.uid", "1002"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.gid", "998"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "certificate_info.mode", "7"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.shell", "/bin/zsh"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.before.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.after.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.renew.on_error.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.shell", "/bin/fish"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.before.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.after.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "hooks.sign.on_error.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "key_info.type", "ECDSA_P384"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "key_info.format", "DEFAULT"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "key_info.pub_file", ""),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "reload_info.method", "DBUS"),
					helper.TestCheckResourceAttr("smallstep_workload.nginx", "reload_info.unit_name", "postgres.service"),

					helper.TestCheckResourceAttr("smallstep_workload.generic", "static_sans.#", "0"),
					helper.TestCheckResourceAttr("smallstep_workload.generic", "device_metadata_key_sans.#", "1"),
					helper.TestCheckResourceAttr("smallstep_workload.generic", "device_metadata_key_sans.0", "internal_host"),
				),
			},
		},
	})
}
