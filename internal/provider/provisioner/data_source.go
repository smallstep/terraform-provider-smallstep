package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource implements data.smallstep_provisioner
type DataSource struct {
	client *v20250101.Client
}

func (a *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = provisionerTypeName
}

// Configure adds the Smallstep API client to the data source.
func (a *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20250101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20250101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	a.client = client
}

func (a *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Model

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nameOrID := config.ID.ValueString()
	if nameOrID == "" {
		nameOrID = config.Name.ValueString()
	}
	httpResp, err := a.client.GetProvisioner(ctx, config.AuthorityID.ValueString(), nameOrID, &v20250101.GetProvisionerParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read provisioner %s: %v", config.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading provisioner %s: %s", reqID, httpResp.StatusCode, nameOrID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	provisioner := &v20250101.Provisioner{}
	if err := json.NewDecoder(httpResp.Body).Decode(provisioner); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal provisioner %s: %v", nameOrID, err),
		)
		return
	}
	state, diags := fromAPI(ctx, provisioner, config.AuthorityID.ValueString(), req.Config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read provisioner %q data source", nameOrID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	prov, provProps, err := utils.Describe("provisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}

	options, _, err := utils.Describe("provisionerOptions")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}

	x509, x509Props, err := utils.Describe("x509Options")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}

	ssh, sshProps, err := utils.Describe("sshOptions")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}

	claims, claimsProps, err := utils.Describe("provisionerClaims")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}

	jwk, jwkProps, err := utils.Describe("jwkProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	// TODO Remove when description is added to OpenAPI spec
	if jwk == "" {
		jwk = "A [provisioner](https://smallstep.com/docs/step-ca/provisioners/#jwk) that uses public-key cryptography to sign and validate a JSON Web Token (JWT)."
	}
	jwk += " This object is populated when type is `JWK`."

	oidc, oidcProps, err := utils.Describe("oidcProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	// TODO Remove when description is added to OpenAPI spec
	if oidc == "" {
		oidc = "A [provisioner](https://smallstep.com/docs/step-ca/provisioners/#oauthoidc-single-sign-on) that trusts an OAuth provider's ID tokens for authentication."
	}
	oidc += " This object is populated when type is `OIDC`."

	acme, acmeProps, err := utils.Describe("acmeProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	acme += " This object is populated when type is `ACME`."

	attest, attestProps, err := utils.Describe("acmeAttestationProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	attest += " This object is populated when type is `ACME_ATTESTATION`."

	x5c, x5cProps, err := utils.Describe("x5cProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	x5c += " This object is populated when type is `X5C`."

	aws, awsProps, err := utils.Describe("awsProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	aws += " This object is populated when type is `AWS`."

	gcp, gcpProps, err := utils.Describe("gcpProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	gcp += " This object is populated when type is `GCP`."

	azure, azureProps, err := utils.Describe("azureProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	azure += " This object is populated when type is `AZURE`."

	resp.Schema = schema.Schema{
		MarkdownDescription: prov,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: provProps["id"] + " Either id or name must be set.",
				Optional:            true,
			},
			"authority_id": schema.StringAttribute{
				// authority_id is a path parameter - it doesn't exist on the
				// provisioner schema in the openAPI spec
				MarkdownDescription: "The UUID of the authority this provisioner is attached to",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: provProps["name"] + " Either id or name must be set.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: provProps["type"],
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: provProps["createdAt"],
				Computed:            true,
			},
			"claims": schema.SingleNestedAttribute{
				MarkdownDescription: claims,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"disable_renewal": schema.BoolAttribute{
						MarkdownDescription: claimsProps["disableRenewal"],
						Computed:            true,
					},
					"allow_renewal_after_expiry": schema.BoolAttribute{
						MarkdownDescription: claimsProps["allowRenewalAfterExpiry"],
						Computed:            true,
					},
					"enable_ssh_ca": schema.BoolAttribute{
						MarkdownDescription: claimsProps["enableSSHCA"],
						Computed:            true,
					},
					"min_tls_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["minTLSCertDuration"],
						Computed:            true,
					},
					"max_tls_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["maxTLSCertDuration"],
						Computed:            true,
					},
					"default_tls_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["defaultTLSCertDuration"],
						Computed:            true,
					},
					"min_user_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["minUserSSHCertDuration"],
						Computed:            true,
					},
					"max_user_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["maxUserSSHCertDuration"],
						Computed:            true,
					},
					"default_user_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["defaultUserSSHCertDuration"],
						Computed:            true,
					},
					"min_host_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["minHostSSHCertDuration"],
						Computed:            true,
					},
					"max_host_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["maxHostSSHCertDuration"],
						Computed:            true,
					},
					"default_host_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["defaultHostSSHCertDuration"],
						Computed:            true,
					},
				},
			},
			"options": schema.SingleNestedAttribute{
				MarkdownDescription: options,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"x509": schema.SingleNestedAttribute{
						MarkdownDescription: x509,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"template": schema.StringAttribute{
								MarkdownDescription: x509Props["template"],
								Computed:            true,
							},
							"template_data": schema.StringAttribute{
								MarkdownDescription: x509Props["templateData"],
								Computed:            true,
							},
						},
					},
					"ssh": schema.SingleNestedAttribute{
						MarkdownDescription: ssh,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"template": schema.StringAttribute{
								MarkdownDescription: sshProps["template"],
								Computed:            true,
							},
							"template_data": schema.StringAttribute{
								MarkdownDescription: sshProps["templateData"],
								Computed:            true,
							},
						},
					},
				},
			},
			"jwk": schema.SingleNestedAttribute{
				MarkdownDescription: jwk,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"key": schema.StringAttribute{
						MarkdownDescription: jwkProps["key"],
						Computed:            true,
					},
					"encrypted_key": schema.StringAttribute{
						MarkdownDescription: jwkProps["encryptedKey"],
						Computed:            true,
					},
				},
			},
			"oidc": schema.SingleNestedAttribute{
				MarkdownDescription: oidc,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						MarkdownDescription: oidcProps["clientID"],
						Computed:            true,
					},
					"client_secret": schema.StringAttribute{
						MarkdownDescription: oidcProps["clientSecret"],
						Computed:            true,
					},
					"configuration_endpoint": schema.StringAttribute{
						MarkdownDescription: oidcProps["configurationEndpoint"],
						Computed:            true,
					},
					"admins": schema.SetAttribute{
						MarkdownDescription: oidcProps["admins"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"domains": schema.SetAttribute{
						MarkdownDescription: oidcProps["domains"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"groups": schema.SetAttribute{
						MarkdownDescription: oidcProps["groups"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"listen_address": schema.StringAttribute{
						MarkdownDescription: oidcProps["listenAddress"],
						Computed:            true,
					},
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: oidcProps["tenantID"],
						Computed:            true,
					},
				},
			},
			"acme": schema.SingleNestedAttribute{
				MarkdownDescription: acme,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"challenges": schema.SetAttribute{
						MarkdownDescription: acmeProps["challenges"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"require_eab": schema.BoolAttribute{
						MarkdownDescription: acmeProps["requireEAB"],
						Computed:            true,
					},
					"force_cn": schema.BoolAttribute{
						MarkdownDescription: acmeProps["forceCN"],
						Computed:            true,
					},
				},
			},
			"acme_attestation": schema.SingleNestedAttribute{
				MarkdownDescription: attest,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"attestation_formats": schema.SetAttribute{
						MarkdownDescription: attestProps["attestationFormats"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"attestation_roots": schema.SetAttribute{
						MarkdownDescription: attestProps["attestationRoots"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"require_eab": schema.BoolAttribute{
						MarkdownDescription: attestProps["requireEAB"],
						Computed:            true,
					},
					"force_cn": schema.BoolAttribute{
						MarkdownDescription: attestProps["forceCN"],
						Computed:            true,
					},
				},
			},
			"x5c": schema.SingleNestedAttribute{
				MarkdownDescription: x5c,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"roots": schema.SetAttribute{
						MarkdownDescription: x5cProps["roots"],
						ElementType:         types.StringType,
						Computed:            true,
					},
				},
			},
			"aws": schema.SingleNestedAttribute{
				MarkdownDescription: aws,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"accounts": schema.SetAttribute{
						MarkdownDescription: awsProps["accounts"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"instance_age": schema.StringAttribute{
						MarkdownDescription: awsProps["instanceAge"],
						Computed:            true,
					},
					"disable_trust_on_first_use": schema.BoolAttribute{
						MarkdownDescription: awsProps["disableTrustOnFirstUse"],
						Computed:            true,
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: awsProps["disableCustomSANs"],
						Computed:            true,
					},
				},
			},
			"gcp": schema.SingleNestedAttribute{
				MarkdownDescription: gcp,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"project_ids": schema.SetAttribute{
						MarkdownDescription: gcpProps["projectIDs"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"service_accounts": schema.SetAttribute{
						MarkdownDescription: gcpProps["serviceAccounts"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"instance_age": schema.StringAttribute{
						MarkdownDescription: gcpProps["instanceAge"],
						Computed:            true,
					},
					"disable_trust_on_first_use": schema.BoolAttribute{
						MarkdownDescription: gcpProps["disableTrustOnFirstUse"],
						Computed:            true,
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: gcpProps["disableCustomSANs"],
						Computed:            true,
					},
				},
			},
			"azure": schema.SingleNestedAttribute{
				MarkdownDescription: azure,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: azureProps["tenantID"],
						Computed:            true,
					},
					"resource_groups": schema.SetAttribute{
						MarkdownDescription: azureProps["resourceGroups"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"audience": schema.StringAttribute{
						MarkdownDescription: azureProps["audience"],
						Computed:            true,
					},
					"disable_trust_on_first_use": schema.BoolAttribute{
						MarkdownDescription: azureProps["disableTrustOnFirstUse"],
						Computed:            true,
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: azureProps["disableCustomSANs"],
						Computed:            true,
					},
				},
			},
		},
	}
}
