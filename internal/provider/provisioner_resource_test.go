package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

func TestAccProvisionerResource(t *testing.T) {
	t.Parallel()

	authority := utils.NewAuthority(t)

	acmeConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "acme" {
	authority_id = %q
	name = "acme foo"
	type = "ACME"
	acme = {
		challenges = ["http-01", "dns-01", "tls-alpn-01"]
		require_eab = true
		force_cn = true
	}
}`, authority.Id)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acmeConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_provisioner.acme", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_provisioner.acme", "type", "ACME"),
					resource.TestCheckResourceAttr("smallstep_provisioner.acme", "name", "acme foo"),
					resource.TestMatchResourceAttr("smallstep_provisioner.acme", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestCheckResourceAttr("smallstep_provisioner.acme", "acme.require_eab", "true"),
				),
			},
		},
	})

	pubJSON, priv := utils.NewJWK(t, "pass")

	jwkConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "jwk" {
	authority_id = %q
	name = "eng@smallstep.com"
	type = "JWK"
	jwk = {
		key = %q
		encrypted_key = %q
	}
}`, authority.Id, pubJSON, priv)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: jwkConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_provisioner.jwk", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_provisioner.jwk", "type", "JWK"),
					resource.TestCheckResourceAttr("smallstep_provisioner.jwk", "name", "eng@smallstep.com"),
					resource.TestMatchResourceAttr("smallstep_provisioner.jwk", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestCheckResourceAttr("smallstep_provisioner.jwk", "jwk.encrypted_key", priv),
					resource.TestCheckResourceAttr("smallstep_provisioner.jwk", "jwk.key", pubJSON),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.jwk",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "eng@smallstep.com"),
				ImportStateVerify: false, // jwk serialized key may be different
			},
		},
	})

	oidcConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "oidc" {
	authority_id = %q
	name = "smallstep.com"
	type = "OIDC"
	oidc = {
		client_id = "abc"
		client_secret = "123"
		configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
		domains = ["smallstep.com"]
		groups = ["eng"]
		admins = ["eng@smallstep.com"]
		listen_address = "localhost:9999"
		tenant_id = "7"
	}
	claims = {
		disable_renewal = true
	}
}`, authority.Id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: oidcConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("smallstep_provisioner.oidc", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "type", "OIDC"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "name", "smallstep.com"),
					resource.TestMatchResourceAttr("smallstep_provisioner.oidc", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.client_id", "abc"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.client_secret", "123"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.configuration_endpoint", "https://accounts.google.com/.well-known/openid-configuration"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.admins.0", "eng@smallstep.com"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.domains.0", "smallstep.com"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.groups.0", "eng"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.listen_address", "localhost:9999"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.tenant_id", "7"),
					resource.TestCheckResourceAttr("smallstep_provisioner.oidc", "claims.disable_renewal", "true"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.oidc",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "smallstep.com"),
				ImportStateVerify: true,
			},
		},
	})

	claimsConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "claims" {
	authority_id = %q
	name = "claims"
	type = "OIDC"
	claims = {
		allow_renewal_after_expiry = true
		enable_ssh_ca = true
		min_tls_cert_duration = "5m"
		max_tls_cert_duration = "45m"
		default_tls_cert_duration = "15m"
		min_user_ssh_cert_duration = "6m"
		max_user_ssh_cert_duration = "46m"
		default_user_ssh_cert_duration = "16m"
		min_host_ssh_cert_duration = "7m"
		max_host_ssh_cert_duration = "47m"
		default_host_ssh_cert_duration = "17m"
	}
	oidc = {
		client_id = "abc"
		client_secret = "123"
		configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
	}
}`, authority.Id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: claimsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.allow_renewal_after_expiry", "true"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.enable_ssh_ca", "true"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.min_tls_cert_duration", "5m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.max_tls_cert_duration", "45m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.default_tls_cert_duration", "15m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.min_user_ssh_cert_duration", "6m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.max_user_ssh_cert_duration", "46m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.default_user_ssh_cert_duration", "16m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.min_host_ssh_cert_duration", "7m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.max_host_ssh_cert_duration", "47m"),
					resource.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.default_host_ssh_cert_duration", "17m"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.claims",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "claims"),
				ImportStateVerify: false, // the duration strings will differ
			},
		},
	})

	optionsConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "options" {
	authority_id = %q
	name = "option"
	type = "OIDC"
	options = {
		x509 = {
			template = "{{.}}"
			template_data = "{\"role\":\"eng\"}"
		}
		ssh = {
			template = "{{ . }}"
			template_data = "{\"role\":\"ops\"}"
		}
	}
	oidc = {
		client_id = "abc"
		client_secret = "123"
		configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
	}
}`, authority.Id)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: optionsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("smallstep_provisioner.options", "options.x509.template", "{{.}}"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.options",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "option"),
				ImportStateVerify: true,
			},
		},
	})
}
