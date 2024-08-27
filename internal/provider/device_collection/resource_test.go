package device_collection

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	helper "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/attestation_authority"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/authority"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
	"github.com/smallstep/terraform-provider-smallstep/internal/testprovider"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	helper.TestMain(m)
}

var provider = &testprovider.SmallstepTestProvider{
	ResourceFactories: []func() resource.Resource{
		NewResource,
		attestation_authority.NewResource,
		authority.NewResource,
	},
}

var providerFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"smallstep": providerserver.NewProtocol6WithError(provider),
}

func TestAccDeviceCollectionResource(t *testing.T) {
	slug := utils.Slug(t)
	awsRequired := fmt.Sprintf(`
resource "smallstep_authority" "agent" {
        name = "Agent"
        subdomain = "agent"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "aws_required_only" {
	slug = %q
	display_name = "EC2 West"
	authority_id = smallstep_authority.agent.id
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
	}
}`, slug)

	updated := fmt.Sprintf(`
resource "smallstep_authority" "agent" {
        name = "Agent"
        subdomain = "agent"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "aws_required_only" {
	slug = %q
	display_name = "EC2 East"
	authority_id = smallstep_authority.agent.id
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsRequired,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "display_name", "EC2 West"),
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "aws_vm.accounts.0", "0123456789"),
					helper.TestMatchResourceAttr("smallstep_device_collection.aws_required_only", "authority_id", utils.UUID),
				),
			},
			{
				Config: updated,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_device_collection.aws_required_only", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_required_only", "display_name", "EC2 East"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	awsOptionalEmpty := fmt.Sprintf(`
resource "smallstep_authority" "aws" {
        name = "AWS"
        subdomain = "aws"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "aws_optional_empty" {
	slug = %q
	display_name = "EC2 West"
	authority_id = smallstep_authority.aws.id
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
		disable_custom_sans = false
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsOptionalEmpty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_optional_empty", "aws_vm.disable_custom_sans", "false"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	awsOptionalNonempty := fmt.Sprintf(`
resource "smallstep_authority" "aws" {
        name = "AWS"
        subdomain = "aws"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "aws_optional_nonempty" {
	slug = %q
	display_name = "EC2 West"
	authority_id = smallstep_authority.aws.id
	device_type = "aws-vm"
	aws_vm = {
		accounts = ["0123456789"]
		disable_custom_sans = true
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: awsOptionalNonempty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.aws_optional_nonempty", "aws_vm.disable_custom_sans", "true"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	gcpRequired := fmt.Sprintf(`
resource "smallstep_authority" "gcp" {
        name = "GCP"
        subdomain = "gcp"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "gcp_required_only" {
	slug = %q
	display_name = "GCE"
	authority_id = smallstep_authority.gcp.id
	device_type = "gcp-vm"
	gcp_vm = {
		service_accounts = ["0123456789"]
		project_ids = ["prod-1"]
	}
}`, slug)

	updatedGCPRequired := fmt.Sprintf(`
resource "smallstep_authority" "gcp" {
        name = "GCP"
        subdomain = "gcp"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "gcp_required_only" {
	slug = %q
	display_name = "Google Compute Engine"
	authority_id = smallstep_authority.gcp.id
	device_type = "gcp-vm"
	gcp_vm = {
		service_accounts = ["0123456789"]
		project_ids = ["prod-1"]
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpRequired,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "display_name", "GCE"),
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "gcp_vm.service_accounts.0", "0123456789"),
				),
			},
			{
				Config: updatedGCPRequired,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_device_collection.gcp_required_only", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_required_only", "display_name", "Google Compute Engine"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	gcpOptionalEmpty := fmt.Sprintf(`
resource "smallstep_authority" "gcp" {
        name = "GCP"
        subdomain = "gcp2"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "gcp_optional_empty" {
	slug = %q
	display_name = "GCE"
	authority_id = smallstep_authority.gcp.id
	device_type = "gcp-vm"
	gcp_vm = {
		service_accounts = ["0123456789"]
		disable_custom_sans = false
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpOptionalEmpty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_optional_empty", "gcp_vm.disable_custom_sans", "false"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	gcpOptionalNonempty := fmt.Sprintf(`
resource "smallstep_authority" "gcp" {
        name = "GCP"
        subdomain = "gcp3"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "gcp_optional_nonempty" {
	slug = %q
	display_name = "GCE"
	authority_id = smallstep_authority.gcp.id
	device_type = "gcp-vm"
	gcp_vm = {
		project_ids = ["prod-123"]
		disable_custom_sans = true
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: gcpOptionalNonempty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.gcp_optional_nonempty", "gcp_vm.disable_custom_sans", "true"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	azureRequired := fmt.Sprintf(`
resource "smallstep_authority" "azure" {
        name = "Azure"
        subdomain = "azure"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "azure_required_only" {
	slug = %q
	display_name = "Azure VMs"
	authority_id = smallstep_authority.azure.id
	device_type = "azure-vm"
	azure_vm = {
		resource_groups = ["0123456789"]
		tenant_id = "7"
	}
}`, slug)

	updatedAzureRequired := fmt.Sprintf(`
resource "smallstep_authority" "azure" {
        name = "Azure"
        subdomain = "azure"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "azure_required_only" {
	slug = %q
	display_name = "Azure Instances"
	authority_id = smallstep_authority.azure.id
	device_type = "azure-vm"
	azure_vm = {
		resource_groups = ["0123456789"]
		tenant_id = "7"
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: azureRequired,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_required_only", "display_name", "Azure VMs"),
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_required_only", "azure_vm.tenant_id", "7"),
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_required_only", "azure_vm.resource_groups.0", "0123456789"),
				),
			},
			{
				Config: updatedAzureRequired,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_device_collection.azure_required_only", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_required_only", "display_name", "Azure Instances"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	azureOptionalEmpty := fmt.Sprintf(`
resource "smallstep_authority" "azure" {
        name = "Azure Agents"
        subdomain = "azure"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "azure_optional_empty" {
	slug = %q
	display_name = "Azure VMs"
	authority_id = smallstep_authority.azure.id
	device_type = "azure-vm"
	azure_vm = {
		tenant_id = "7"
		resource_groups = ["0123456789"]
		disable_custom_sans = false
		audience = ""
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: azureOptionalEmpty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_optional_empty", "azure_vm.disable_custom_sans", "false"),
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_optional_empty", "azure_vm.audience", ""),
				),
			},
		},
	})

	slug = utils.Slug(t)
	azureOptionalNonempty := fmt.Sprintf(`
resource "smallstep_authority" "azure" {
        name = "Azure Agents"
        subdomain = "azure2"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "azure_optional_nonempty" {
	slug = %q
	display_name = "Azure VMs"
	authority_id = smallstep_authority.azure.id
	device_type = "azure-vm"
	azure_vm = {
		tenant_id = "7"
		resource_groups = ["0123456789"]
		disable_custom_sans = true
		audience = "example.com"
	}
}`, slug)

	helper.Test(t, helper.TestCase{
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: azureOptionalNonempty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_optional_nonempty", "azure_vm.disable_custom_sans", "true"),
					helper.TestCheckResourceAttr("smallstep_device_collection.azure_optional_nonempty", "azure_vm.audience", "example.com"),
				),
			},
		},
	})

	attestorRoot, attestorIntermediate := utils.CACerts(t)
	slug = utils.Slug(t)
	tpmRequired := fmt.Sprintf(`
resource "smallstep_attestation_authority" "attest_ca" {
	name = %q
	attestor_roots = %q
}

resource "smallstep_authority" "linux_agents" {
        name = "Linux Agents"
        subdomain = "linux"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "tpm_required_only" {
	slug = %q
	display_name = "TPMs"
	authority_id = smallstep_authority.linux_agents.id
	device_type = "tpm"
	tpm = {}
	depends_on = [smallstep_attestation_authority.attest_ca]
}`, slug, attestorRoot, slug)

	tpmUpdated := fmt.Sprintf(`
resource "smallstep_attestation_authority" "attest_ca" {
	name = %q
	attestor_roots = %q
}

resource "smallstep_authority" "linux_agents" {
        name = "Linux Agents"
        subdomain = "linux"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "tpm_required_only" {
	slug = %q
	display_name = "TPM Servers"
	authority_id = smallstep_authority.linux_agents.id
	device_type = "tpm"
	tpm = {
		require_eab = true
		force_cn = true
	}
	depends_on = [smallstep_attestation_authority.attest_ca]
}`, slug, attestorRoot, slug)

	helper.Test(t, helper.TestCase{
		PreCheck: func() {
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: tpmRequired,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_required_only", "display_name", "TPMs"),
				),
			},
			{
				Config: tpmUpdated,
				ConfigPlanChecks: helper.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("smallstep_device_collection.tpm_required_only", plancheck.ResourceActionUpdate),
					},
				},
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_required_only", "slug", slug),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_required_only", "display_name", "TPM Servers"),
				),
			},
		},
	})

	slug = utils.Slug(t)
	tpmOptionalEmpty := fmt.Sprintf(`
resource "smallstep_attestation_authority" "attest_ca" {
	name = %q
	attestor_roots = %q
}

resource "smallstep_authority" "linux_agents" {
        name = "Linux Agents"
        subdomain = "linux2"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "tpm_optional_empty" {
	slug = %q
	display_name = "TPM Servers"
	authority_id = smallstep_authority.linux_agents.id
	device_type = "tpm"
	tpm = {
		force_cn = false
		require_eab = false
		attestor_roots = ""
		attestor_intermediates = ""
	}
	depends_on = [smallstep_attestation_authority.attest_ca]
}`, slug, attestorRoot, slug)

	helper.Test(t, helper.TestCase{
		PreCheck: func() {
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: tpmOptionalEmpty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_empty", "tpm.force_cn", "false"),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_empty", "tpm.require_eab", "false"),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_empty", "tpm.attestor_roots", ""),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_empty", "tpm.attestor_intermediates", ""),
				),
			},
		},
	})

	slug = utils.Slug(t)
	tpmOptionalNonempty := fmt.Sprintf(`
resource "smallstep_authority" "linux_agents" {
        name = "Linux Agents"
        subdomain = "linux3"
        type = "devops"
        admin_emails = ["andrew@smallstep.com"]
}

resource "smallstep_device_collection" "tpm_optional_nonempty" {
			slug = %q
			display_name = "TPM Servers"
			authority_id = smallstep_authority.linux_agents.id
			device_type = "tpm"
			tpm = {
				attestor_roots = %q
				attestor_intermediates = %q
				force_cn = true
				require_eab = true
			}
		}`, slug, attestorRoot, attestorIntermediate)

	helper.Test(t, helper.TestCase{
		PreCheck: func() {
			require.NoError(t, utils.SweepAttestationAuthorities())
		},
		ProtoV6ProviderFactories: providerFactories,
		Steps: []helper.TestStep{
			{
				Config: tpmOptionalNonempty,
				Check: helper.ComposeAggregateTestCheckFunc(
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_nonempty", "tpm.force_cn", "true"),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_nonempty", "tpm.require_eab", "true"),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_nonempty", "tpm.attestor_roots", attestorRoot),
					helper.TestCheckResourceAttr("smallstep_device_collection.tpm_optional_nonempty", "tpm.attestor_intermediates", attestorIntermediate),
				),
			},
		},
	})
}
