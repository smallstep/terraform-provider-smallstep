package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the provisioner resource implementation.
type Resource struct {
	client *v20250101.Client
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = provisionerTypeName
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
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
	jwk += " This object is required when type is `JWK` and is otherwise ignored."

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
	oidc += " This object is required when type is `OIDC` and is otherwise ignored."

	acme, acmeProps, err := utils.Describe("acmeProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	acme += " This object is required when type is `ACME` and is otherwise ignored."

	attest, attestProps, err := utils.Describe("acmeAttestationProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	attest += " This object is required when type is `ACME_ATTESTATION` and is otherwise ignored."

	x5c, x5cProps, err := utils.Describe("x5cProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	x5c += " This object is required when type is `X5C` and is otherwise ignored."

	aws, awsProps, err := utils.Describe("awsProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	aws += " This object is required when type is `AWS` and is otherwise ignored."

	gcp, gcpProps, err := utils.Describe("gcpProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	gcp += " This object is required when type is `GCP` and is otherwise ignored."

	azure, azureProps, err := utils.Describe("azureProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}
	azure += " This object is required when type is `AZURE` and is otherwise ignored."

	resp.Schema = schema.Schema{
		MarkdownDescription: prov,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: provProps["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"authority_id": schema.StringAttribute{
				// authority_id is a path parameter - it doesn't exist on the
				// provisioner schema in the openAPI spec
				MarkdownDescription: "The UUID of the authority this provisioner is attached to",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: provProps["name"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: provProps["type"],
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: provProps["createdAt"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"claims": schema.SingleNestedAttribute{
				MarkdownDescription: claims,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"disable_renewal": schema.BoolAttribute{
						MarkdownDescription: claimsProps["disableRenewal"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"allow_renewal_after_expiry": schema.BoolAttribute{
						MarkdownDescription: claimsProps["allowRenewalAfterExpiry"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"enable_ssh_ca": schema.BoolAttribute{
						MarkdownDescription: claimsProps["enableSSHCA"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"min_tls_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["minTLSCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"max_tls_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["maxTLSCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"default_tls_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["defaultTLSCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"min_user_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["minUserSSHCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"max_user_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["maxUserSSHCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"default_user_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["defaultUserSSHCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"min_host_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["minHostSSHCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"max_host_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["maxHostSSHCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"default_host_ssh_cert_duration": schema.StringAttribute{
						MarkdownDescription: claimsProps["defaultHostSSHCertDuration"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"options": schema.SingleNestedAttribute{
				MarkdownDescription: options,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"x509": schema.SingleNestedAttribute{
						MarkdownDescription: x509,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"template": schema.StringAttribute{
								MarkdownDescription: x509Props["template"],
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"template_data": schema.StringAttribute{
								MarkdownDescription: x509Props["templateData"],
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
					"ssh": schema.SingleNestedAttribute{
						MarkdownDescription: ssh,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"template": schema.StringAttribute{
								MarkdownDescription: sshProps["template"],
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
							"template_data": schema.StringAttribute{
								MarkdownDescription: sshProps["templateData"],
								Optional:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.RequiresReplace(),
								},
							},
						},
					},
				},
			},
			"jwk": schema.SingleNestedAttribute{
				MarkdownDescription: jwk,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"key": schema.StringAttribute{
						MarkdownDescription: jwkProps["key"],
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"encrypted_key": schema.StringAttribute{
						MarkdownDescription: jwkProps["encryptedKey"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"oidc": schema.SingleNestedAttribute{
				MarkdownDescription: oidc,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						MarkdownDescription: oidcProps["clientID"],
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"client_secret": schema.StringAttribute{
						MarkdownDescription: oidcProps["clientSecret"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"configuration_endpoint": schema.StringAttribute{
						MarkdownDescription: oidcProps["configurationEndpoint"],
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"admins": schema.SetAttribute{
						MarkdownDescription: oidcProps["admins"],
						ElementType:         types.StringType,
						Optional:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"domains": schema.SetAttribute{
						MarkdownDescription: oidcProps["domains"],
						ElementType:         types.StringType,
						Optional:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"groups": schema.SetAttribute{
						MarkdownDescription: oidcProps["groups"],
						ElementType:         types.StringType,
						Optional:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"listen_address": schema.StringAttribute{
						MarkdownDescription: oidcProps["listenAddress"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: oidcProps["tenantID"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"acme": schema.SingleNestedAttribute{
				MarkdownDescription: acme,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"challenges": schema.SetAttribute{
						MarkdownDescription: acmeProps["challenges"],
						ElementType:         types.StringType,
						Required:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"require_eab": schema.BoolAttribute{
						MarkdownDescription: acmeProps["requireEAB"],
						Required:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"force_cn": schema.BoolAttribute{
						MarkdownDescription: acmeProps["forceCN"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
						Default: booldefault.StaticBool(false),
					},
				},
			},
			"acme_attestation": schema.SingleNestedAttribute{
				MarkdownDescription: attest,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"attestation_formats": schema.SetAttribute{
						MarkdownDescription: attestProps["attestationFormats"],
						ElementType:         types.StringType,
						Required:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"attestation_roots": schema.SetAttribute{
						MarkdownDescription: attestProps["attestationRoots"],
						ElementType:         types.StringType,
						Optional:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"require_eab": schema.BoolAttribute{
						MarkdownDescription: attestProps["requireEAB"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
						Default: booldefault.StaticBool(false),
					},
					"force_cn": schema.BoolAttribute{
						MarkdownDescription: attestProps["forceCN"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
						Default: booldefault.StaticBool(false),
					},
				},
			},
			"x5c": schema.SingleNestedAttribute{
				MarkdownDescription: x5c,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"roots": schema.SetAttribute{
						MarkdownDescription: x5cProps["roots"],
						Required:            true,
						ElementType:         types.StringType,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"aws": schema.SingleNestedAttribute{
				MarkdownDescription: aws,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"accounts": schema.SetAttribute{
						MarkdownDescription: awsProps["accounts"],
						ElementType:         types.StringType,
						Required:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"instance_age": schema.StringAttribute{
						MarkdownDescription: awsProps["instanceAge"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"disable_trust_on_first_use": schema.BoolAttribute{
						MarkdownDescription: awsProps["disableTrustOnFirstUse"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: awsProps["disableCustomSANs"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"gcp": schema.SingleNestedAttribute{
				MarkdownDescription: gcp,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"project_ids": schema.SetAttribute{
						MarkdownDescription: gcpProps["projectIDs"],
						ElementType:         types.StringType,
						Required:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"service_accounts": schema.SetAttribute{
						MarkdownDescription: gcpProps["serviceAccounts"],
						ElementType:         types.StringType,
						Required:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"instance_age": schema.StringAttribute{
						MarkdownDescription: gcpProps["instanceAge"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"disable_trust_on_first_use": schema.BoolAttribute{
						MarkdownDescription: gcpProps["disableTrustOnFirstUse"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: gcpProps["disableCustomSANs"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
				},
			},
			"azure": schema.SingleNestedAttribute{
				MarkdownDescription: azure,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"tenant_id": schema.StringAttribute{
						MarkdownDescription: azureProps["tenantID"],
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"resource_groups": schema.SetAttribute{
						MarkdownDescription: azureProps["resourceGroups"],
						ElementType:         types.StringType,
						Required:            true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.RequiresReplace(),
						},
					},
					"audience": schema.StringAttribute{
						MarkdownDescription: azureProps["audience"],
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"disable_trust_on_first_use": schema.BoolAttribute{
						MarkdownDescription: azureProps["disableTrustOnFirstUse"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
					"disable_custom_sans": schema.BoolAttribute{
						MarkdownDescription: azureProps["disableCustomSANs"],
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20250101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *v20250101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Model

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	p, err := toAPI(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client",
			err.Error(),
		)
		return
	}

	b, _ := json.Marshal(p)
	tflog.Trace(ctx, string(b))

	httpResp, err := a.client.PostAuthorityProvisioners(ctx, plan.AuthorityID.ValueString(), &v20250101.PostAuthorityProvisionersParams{}, *p)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create provisioner %q: %v", plan.Name.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	provisioner := &v20250101.Provisioner{}
	if err := json.NewDecoder(httpResp.Body).Decode(provisioner); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal provisioner %q: %v", plan.Name.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(ctx, provisioner, plan.AuthorityID.ValueString(), req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Use the original key if it's a jwk provisioner. Terraform will fail if
	// the state does not match the plan but there are many valid json
	// serializations of the key
	if plan.JWK != nil && state.JWK != nil && utils.IsJSONEqual(plan.JWK.Key.ValueString(), state.JWK.Key.ValueString()) {
		state.JWK.Key = plan.JWK.Key
	}
	// A null claims is returned as an empty claims object. Restore it to null
	// to prevent errors from terraform.
	if plan.Claims == nil && state.Claims != nil && state.Claims.isEmpty() {
		state.Claims = nil
	}
	// Durations such as "1h" are returned as "1h0m0s". Terraform will fail if
	// the state does not match the plan so use the state value if equal.
	useConfiguredDurationIfEqual(plan.Claims, state.Claims)

	tflog.Trace(ctx, fmt.Sprintf("create provisioner %q resource", plan.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (a *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &Model{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nameOrID := state.ID.ValueString()
	if nameOrID == "" {
		nameOrID = state.Name.ValueString()
	}
	httpResp, err := a.client.GetProvisioner(ctx, state.AuthorityID.ValueString(), nameOrID, &v20250101.GetProvisionerParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read provisioner %q: %v", state.ID, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d getting provisioner %s: %s", reqID, httpResp.StatusCode, nameOrID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	provisioner := &v20250101.Provisioner{}
	if err := json.NewDecoder(httpResp.Body).Decode(provisioner); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal provisioner %q: %v", state.ID.ValueString(), err),
		)
		return
	}

	actual, diags := fromAPI(ctx, provisioner, state.AuthorityID.ValueString(), req.State)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Use the original key if it's a jwk provisioner. Terraform will fail if
	// the stored state does not match the current state but there are many valid
	// json serializations of the key.
	if state.JWK != nil && actual.JWK != nil && utils.IsJSONEqual(state.JWK.Key.ValueString(), actual.JWK.Key.ValueString()) {
		actual.JWK.Key = state.JWK.Key
	}

	useConfiguredDurationIfEqual(state.Claims, actual.Claims)

	// A null claims is returned as an empty claims object. Restore it to null
	// to prevent errors from terraform.
	if state.Claims == nil && actual.Claims != nil && actual.Claims.isEmpty() {
		actual.Claims = nil
	}

	tflog.Trace(ctx, fmt.Sprintf("read provisioner %q resource", state.ID.ValueString()))

	resp.Diagnostics.Append(resp.State.Set(ctx, actual)...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Update not supported. All changes require replacement.
}

func (a *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Model

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	nameOrID := state.ID.ValueString()
	if nameOrID == "" {
		nameOrID = state.Name.ValueString()
	}
	httpResp, err := a.client.DeleteProvisioner(ctx, state.AuthorityID.ValueString(), nameOrID, &v20250101.DeleteProvisionerParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete provisioner %s: %v", state.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting provisioner %s: %s", reqID, httpResp.StatusCode, state.ID.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Import ID must be "<authority_id>/<name>"`,
		)
		return
	}
	if _, err := uuid.Parse(parts[0]); err != nil {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Import authority_id is not a valid UUID"`,
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), "")...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("authority_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}

func durationEqual(a, b types.String) bool {
	if a.IsNull() || b.IsNull() {
		return false
	}
	aDuration, err := time.ParseDuration(a.ValueString())
	if err != nil {
		return false
	}
	bDuration, err := time.ParseDuration(b.ValueString())
	if err != nil {
		return false
	}

	return aDuration == bDuration
}

// Duration such as "1h" are returned as "1h0m0s". Terraform fails because the
// strings do not match.
func useConfiguredDurationIfEqual(tf, api *ClaimsModel) {
	if tf == nil || api == nil {
		return
	}

	if durationEqual(tf.MinTLSCertDuration, tf.MinTLSCertDuration) {
		api.MinTLSCertDuration = tf.MinTLSCertDuration
	}
	if durationEqual(tf.MaxTLSCertDuration, api.MaxTLSCertDuration) {
		api.MaxTLSCertDuration = tf.MaxTLSCertDuration
	}
	if durationEqual(tf.DefaultTLSCertDuration, api.DefaultTLSCertDuration) {
		api.DefaultTLSCertDuration = tf.DefaultTLSCertDuration
	}
	if durationEqual(tf.MinUserSSHCertDuration, api.MinUserSSHCertDuration) {
		api.MinUserSSHCertDuration = tf.MinUserSSHCertDuration
	}
	if durationEqual(tf.MaxUserSSHCertDuration, api.MaxUserSSHCertDuration) {
		api.MaxUserSSHCertDuration = tf.MaxUserSSHCertDuration
	}
	if durationEqual(tf.DefaultUserSSHCertDuration, api.DefaultUserSSHCertDuration) {
		api.DefaultUserSSHCertDuration = tf.DefaultUserSSHCertDuration
	}
	if durationEqual(tf.MinHostSSHCertDuration, api.MinHostSSHCertDuration) {
		api.MinHostSSHCertDuration = tf.MinHostSSHCertDuration
	}
	if durationEqual(tf.MaxHostSSHCertDuration, api.MaxHostSSHCertDuration) {
		api.MaxHostSSHCertDuration = tf.MaxHostSSHCertDuration
	}
	if durationEqual(tf.DefaultHostSSHCertDuration, api.DefaultHostSSHCertDuration) {
		api.DefaultHostSSHCertDuration = tf.DefaultHostSSHCertDuration
	}
}
