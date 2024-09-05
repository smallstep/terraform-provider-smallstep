package provisioner

import (
	"fmt"
	"regexp"
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

func TestAccProvisionerResource(t *testing.T) {
	t.Parallel()

	authority := utils.NewAuthority(t)

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

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: jwkConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.jwk", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk", "type", "JWK"),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk", "name", "eng@smallstep.com"),
					helper.TestMatchResourceAttr("smallstep_provisioner.jwk", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk", "jwk.encrypted_key", priv),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk", "jwk.key", pubJSON),
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

	jwkEmptyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "jwk_empty" {
			authority_id = %q
			name = "jwk empty"
			type = "JWK"
			jwk = {
				key = %q
				encrypted_key = ""
			}
		}`, authority.Id, pubJSON)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: jwkEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.jwk_empty", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk_empty", "type", "JWK"),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk_empty", "name", "jwk empty"),
					helper.TestMatchResourceAttr("smallstep_provisioner.jwk_empty", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk_empty", "jwk.key", pubJSON),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk_empty", "jwk.encrypted_key", ""),
				),
			},
		},
	})

	jwkOnlyRequiredConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "jwk_only_required" {
			authority_id = %q
			name = "jwk only required"
			type = "JWK"
			jwk = {
				key = %q
			}
		}`, authority.Id, pubJSON)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: jwkOnlyRequiredConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.jwk_only_required", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk_only_required", "type", "JWK"),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk_only_required", "name", "jwk only required"),
					helper.TestMatchResourceAttr("smallstep_provisioner.jwk_only_required", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.jwk_only_required", "jwk.key", pubJSON),
				),
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

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: oidcConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.oidc", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "type", "OIDC"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "name", "smallstep.com"),
					helper.TestMatchResourceAttr("smallstep_provisioner.oidc", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.client_id", "abc"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.client_secret", "123"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.configuration_endpoint", "https://accounts.google.com/.well-known/openid-configuration"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.admins.0", "eng@smallstep.com"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.domains.0", "smallstep.com"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.groups.0", "eng"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.listen_address", "localhost:9999"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "oidc.tenant_id", "7"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc", "claims.disable_renewal", "true"),
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

	oidcEmptyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "oidc_empty" {
			authority_id = %q
			name = "empty oidc"
			type = "OIDC"
			oidc = {
				client_id = "abc"
				client_secret = ""
				configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
				domains = []
				groups = []
				admins = []
				listen_address = ""
				tenant_id = ""
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: oidcEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.oidc_empty", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "type", "OIDC"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "name", "empty oidc"),
					helper.TestMatchResourceAttr("smallstep_provisioner.oidc_empty", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.client_id", "abc"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.client_secret", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.configuration_endpoint", "https://accounts.google.com/.well-known/openid-configuration"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.admins.#", "0"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.domains.#", "0"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.groups.#", "0"),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.listen_address", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.oidc_empty", "oidc.tenant_id", ""),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.oidc_empty",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "empty oidc"),
				ImportStateVerify: false,
			},
		},
	})

	// Also tests OIDC with no optional values set
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
				configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: claimsConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.allow_renewal_after_expiry", "true"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.enable_ssh_ca", "true"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.min_tls_cert_duration", "5m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.max_tls_cert_duration", "45m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.default_tls_cert_duration", "15m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.min_user_ssh_cert_duration", "6m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.max_user_ssh_cert_duration", "46m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.default_user_ssh_cert_duration", "16m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.min_host_ssh_cert_duration", "7m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.max_host_ssh_cert_duration", "47m"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims", "claims.default_host_ssh_cert_duration", "17m"),
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

	claimsEmptyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "claims_empty" {
			authority_id = %q
			name = "claims empty"
			type = "OIDC"
			claims = {
				allow_renewal_after_expiry = false
				enable_ssh_ca = false
				disable_renewal = false
				min_tls_cert_duration = ""
				max_tls_cert_duration = ""
				default_tls_cert_duration = ""
				min_user_ssh_cert_duration = ""
				max_user_ssh_cert_duration = ""
				default_user_ssh_cert_duration = ""
				min_host_ssh_cert_duration = ""
				max_host_ssh_cert_duration = ""
				default_host_ssh_cert_duration = ""
			}
			oidc = {
				client_id = "abc"
				configuration_endpoint = "https://accounts.google.com/.well-known/openid-configuration"
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: claimsEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.allow_renewal_after_expiry", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.enable_ssh_ca", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.disable_renewal", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.min_tls_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.max_tls_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.default_tls_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.min_user_ssh_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.max_user_ssh_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.default_user_ssh_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.min_host_ssh_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.max_host_ssh_cert_duration", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.claims_empty", "claims.default_host_ssh_cert_duration", ""),
				),
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

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: optionsConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.options", "options.x509.template", "{{.}}"),
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
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: acmeConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.acme", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme", "type", "ACME"),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme", "name", "acme foo"),
					helper.TestMatchResourceAttr("smallstep_provisioner.acme", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme", "acme.require_eab", "true"),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme", "acme.require_eab", "true"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.acme",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "acme foo"),
				ImportStateVerify: true,
			},
		},
	})

	acmeEmptyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "acme_empty" {
			authority_id = %q
			name = "acme empty"
			type = "ACME"
			acme = {
				challenges = ["http-01"]
				require_eab = true
				force_cn = false
			}
		}`, authority.Id)
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: acmeEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.acme_empty", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme_empty", "type", "ACME"),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme_empty", "name", "acme empty"),
					helper.TestMatchResourceAttr("smallstep_provisioner.acme_empty", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme_empty", "acme.require_eab", "true"),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme_empty", "acme.force_cn", "false"),
				),
			},
		},
	})

	acmeRequiredOnlyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "acme_required_only" {
			authority_id = %q
			name = "acme required only"
			type = "ACME"
			acme = {
				challenges = ["tls-alpn-01", "dns-01", "http-01"]
				require_eab = true
			}
		}`, authority.Id)
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: acmeRequiredOnlyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.acme_required_only", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme_required_only", "type", "ACME"),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme_required_only", "name", "acme required only"),
					helper.TestMatchResourceAttr("smallstep_provisioner.acme_required_only", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.acme_required_only", "acme.require_eab", "true"),
				),
			},
		},
	})

	root, _ := utils.CACerts(t)
	attestConfig := fmt.Sprintf(`
				resource "smallstep_provisioner" "attest" {
					authority_id = %q
					name = "attest foo"
					type = "ACME_ATTESTATION"
					acme_attestation = {
						attestation_formats = ["apple", "step", "tpm"]
						attestation_roots = [%q]
					}
				}`, authority.Id, root)
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: attestConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.attest", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest", "type", "ACME_ATTESTATION"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest", "name", "attest foo"),
					helper.TestMatchResourceAttr("smallstep_provisioner.attest", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest", "acme_attestation.require_eab", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest", "acme_attestation.force_cn", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest", "acme_attestation.attestation_formats.#", "3"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest", "acme_attestation.attestation_roots.0", root),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.attest",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "attest foo"),
				ImportStateVerify: true,
			},
		},
	})

	attestEmptyConfig := fmt.Sprintf(`
				resource "smallstep_provisioner" "attest_empty" {
					authority_id = %q
					name = "attest empty"
					type = "ACME_ATTESTATION"
					acme_attestation = {
						attestation_formats = ["step"]
						attestation_roots = []
					}
				}`, authority.Id)
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: attestEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.attest_empty", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_empty", "type", "ACME_ATTESTATION"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_empty", "name", "attest empty"),
					helper.TestMatchResourceAttr("smallstep_provisioner.attest_empty", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_empty", "acme_attestation.require_eab", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_empty", "acme_attestation.force_cn", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_empty", "acme_attestation.attestation_roots.#", "0"),
				),
			},
		},
	})

	attestOnlyRequiredConfig := fmt.Sprintf(`
				resource "smallstep_provisioner" "attest_only_required" {
					authority_id = %q
					name = "attest only required"
					type = "ACME_ATTESTATION"
					acme_attestation = {
						attestation_formats = ["step", "apple"]
					}
				}`, authority.Id)
	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: attestOnlyRequiredConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.attest_only_required", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_only_required", "type", "ACME_ATTESTATION"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_only_required", "name", "attest only required"),
					helper.TestMatchResourceAttr("smallstep_provisioner.attest_only_required", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_only_required", "acme_attestation.require_eab", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_only_required", "acme_attestation.force_cn", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_only_required", "acme_attestation.attestation_formats.#", "2"),
					helper.TestCheckResourceAttr("smallstep_provisioner.attest_only_required", "acme_attestation.attestation_roots.#", "0"),
				),
			},
		},
	})

	x5cConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "x5c" {
			authority_id = %q
			name = "x5c foo"
			type = "X5C"
			x5c = {
				roots = [%q]
			}
		}`, authority.Id, root)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: x5cConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.x5c", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.x5c", "type", "X5C"),
					helper.TestCheckResourceAttr("smallstep_provisioner.x5c", "name", "x5c foo"),
					helper.TestMatchResourceAttr("smallstep_provisioner.x5c", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.x5c", "x5c.roots.#", "1"),
					helper.TestCheckResourceAttr("smallstep_provisioner.x5c", "x5c.roots.0", root),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.x5c",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "x5c foo"),
				ImportStateVerify: true,
			},
		},
	})

	awsFullConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "aws_full" {
			authority_id = %q
			name = "aws full"
			type = "AWS"
			aws = {
				accounts = ["0123456789"]
				instance_age = "1h"
				disable_trust_on_first_use = true
				disable_custom_sans = true
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsFullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.aws_full", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_full", "type", "AWS"),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_full", "name", "aws full"),
					helper.TestMatchResourceAttr("smallstep_provisioner.aws_full", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_full", "aws.accounts.#", "1"),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_full", "aws.accounts.0", "0123456789"),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_full", "aws.instance_age", "1h"),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_full", "aws.disable_trust_on_first_use", "true"),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_full", "aws.disable_custom_sans", "true"),
				),
			},
			{
				ResourceName:            "smallstep_provisioner.aws_full",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s", authority.Id, "aws full"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"aws.instance_age"},
			},
		},
	})

	awsEmptyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "aws_empty" {
			authority_id = %q
			name = "aws empty"
			type = "AWS"
			aws = {
				accounts = ["0123456789"]
				instance_age = ""
				disable_trust_on_first_use = false
				disable_custom_sans = false
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_empty", "aws.instance_age", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_empty", "aws.disable_trust_on_first_use", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_empty", "aws.disable_custom_sans", "false"),
				),
			},
			{
				ResourceName:  "smallstep_provisioner.aws_empty",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/%s", authority.Id, "aws empty"),
				// empty fields will be null
				ImportStateVerify: false,
			},
		},
	})

	awsNullConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "aws_null" {
			authority_id = %q
			name = "aws null"
			type = "AWS"
			aws = {
				accounts = ["0123456789"]
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsNullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.aws_null", "name", "aws null"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.aws_null",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "aws null"),
				ImportStateVerify: true,
			},
		},
	})

	gcpFullConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "gcp_full" {
	authority_id = %q
	name = "gcp full"
	type = "GCP"
	gcp = {
		project_ids = ["prod-1234"]
		service_accounts = ["pki@prod-1234.iam.gserviceaccount.com"]
		instance_age = "1h"
		disable_trust_on_first_use = true
		disable_custom_sans = true
	}
}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpFullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.gcp_full", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "type", "GCP"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "name", "gcp full"),
					helper.TestMatchResourceAttr("smallstep_provisioner.gcp_full", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "gcp.project_ids.#", "1"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "gcp.project_ids.0", "prod-1234"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "gcp.service_accounts.#", "1"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "gcp.service_accounts.0", "pki@prod-1234.iam.gserviceaccount.com"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "gcp.instance_age", "1h"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "gcp.disable_trust_on_first_use", "true"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_full", "gcp.disable_custom_sans", "true"),
				),
			},
			{
				ResourceName:            "smallstep_provisioner.gcp_full",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s", authority.Id, "gcp full"),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"gcp.instance_age"},
			},
		},
	})

	gcpEmptyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "gcp_empty" {
			authority_id = %q
			name = "gcp empty"
			type = "GCP"
			gcp = {
				project_ids = ["prod-1234"]
				service_accounts = ["pki@prod-1234.iam.gserviceaccount.com"]
				instance_age = ""
				disable_trust_on_first_use = false
				disable_custom_sans = false
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_empty", "gcp.instance_age", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_empty", "gcp.disable_trust_on_first_use", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_empty", "gcp.disable_custom_sans", "false"),
				),
			},
			{
				ResourceName:  "smallstep_provisioner.gcp_empty",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/%s", authority.Id, "gcp empty"),
				// empty fields will be null on import
				ImportStateVerify: false,
			},
		},
	})

	gcpNullConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "gcp_null" {
			authority_id = %q
			name = "gcp null"
			type = "GCP"
			gcp = {
				project_ids = ["prod-1234"]
				service_accounts = ["pki@prod-1234.iam.gserviceaccount.com"]
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpNullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.gcp_null", "name", "gcp null"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.gcp_null",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "gcp null"),
				ImportStateVerify: true,
			},
		},
	})

	azureFullConfig := fmt.Sprintf(`
resource "smallstep_provisioner" "azure_full" {
	authority_id = %q
	name = "azure full"
	type = "AZURE"
	azure = {
		tenant_id = "948920a7-8fc1-431f-be94-030a80232e51"
		resource_groups = ["prod-1234"]
		audience = "example.com"
		disable_trust_on_first_use = true
		disable_custom_sans = true
	}
}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: azureFullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestMatchResourceAttr("smallstep_provisioner.azure_full", "id", regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "type", "AZURE"),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "name", "azure full"),
					helper.TestMatchResourceAttr("smallstep_provisioner.azure_full", "created_at", regexp.MustCompile(`^20\d\d-\d\d-\d\dT\d\d:\d\d:\d\dZ`)),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "azure.resource_groups.#", "1"),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "azure.resource_groups.0", "prod-1234"),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "azure.tenant_id", "948920a7-8fc1-431f-be94-030a80232e51"),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "azure.audience", "example.com"),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "azure.disable_trust_on_first_use", "true"),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_full", "azure.disable_custom_sans", "true"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.azure_full",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "azure full"),
				ImportStateVerify: true,
			},
		},
	})

	azureEmptyConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "azure_empty" {
			authority_id = %q
			name = "azure empty"
			type = "AZURE"
			azure = {
				tenant_id = "948920a7-8fc1-431f-be94-030a80232e51"
				resource_groups = ["prod-1234"]
				audience = ""
				disable_trust_on_first_use = false
				disable_custom_sans = false
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: azureEmptyConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_empty", "azure.audience", ""),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_empty", "azure.disable_trust_on_first_use", "false"),
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_empty", "azure.disable_custom_sans", "false"),
				),
			},
			{
				ResourceName:  "smallstep_provisioner.azure_empty",
				ImportState:   true,
				ImportStateId: fmt.Sprintf("%s/%s", authority.Id, "azure empty"),
				// empty fields will be null on import
				ImportStateVerify: false,
			},
		},
	})

	azureNullConfig := fmt.Sprintf(`
		resource "smallstep_provisioner" "azure_null" {
			authority_id = %q
			name = "azure null"
			type = "AZURE"
			azure = {
				tenant_id = "948920a7-8fc1-431f-be94-030a80232e51"
				resource_groups = ["prod-1234"]
			}
		}`, authority.Id)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: azureNullConfig,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_provisioner.azure_null", "name", "azure null"),
				),
			},
			{
				ResourceName:      "smallstep_provisioner.azure_null",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", authority.Id, "azure null"),
				ImportStateVerify: true,
			},
		},
	})
}
