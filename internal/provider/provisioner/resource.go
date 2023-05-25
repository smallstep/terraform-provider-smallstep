package provisioner

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20230301 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20230301"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the provisioner resource implementation.
type Resource struct {
	client *v20230301.Client
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

	oidc, oidcProps, err := utils.Describe("oidcProvisioner")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI schema",
			err.Error(),
		)
		return
	}

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
								Required:            true,
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
								Required:            true,
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
						MarkdownDescription: jwkProps["encrypted_key"],
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
					"admins": schema.ListAttribute{
						MarkdownDescription: oidcProps["admins"],
						ElementType:         types.StringType,
						Optional:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
					"domains": schema.ListAttribute{
						MarkdownDescription: oidcProps["domains"],
						ElementType:         types.StringType,
						Optional:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
						},
					},
					"groups": schema.ListAttribute{
						MarkdownDescription: oidcProps["groups"],
						ElementType:         types.StringType,
						Optional:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.RequiresReplace(),
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
		},
	}
}

func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20230301.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *v20230301.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	httpResp, err := a.client.PostAuthorityProvisioners(ctx, plan.AuthorityID.ValueString(), &v20230301.PostAuthorityProvisionersParams{}, *p)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create provisioner %q: %v", plan.Name.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d: %s", httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	provisioner := &v20230301.Provisioner{}
	if err := json.NewDecoder(httpResp.Body).Decode(provisioner); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal provisioner %q: %v", plan.Name.ValueString(), err),
		)
		return
	}

	state, diags := fromAPI(provisioner, plan.AuthorityID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Use the original key if it's a jwk provisioner. Terraform will fail if
	// the state does not match the plan but there are many valid json
	// serializations of the key
	if plan.JWK != nil && state.JWK != nil && jsonEqual(plan.JWK.Key.ValueString(), state.JWK.Key.ValueString()) {
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
	httpResp, err := a.client.GetProvisioner(ctx, state.AuthorityID.ValueString(), nameOrID, &v20230301.GetProvisionerParams{})
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
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d getting provisioner %s: %s", httpResp.StatusCode, nameOrID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	provisioner := &v20230301.Provisioner{}
	if err := json.NewDecoder(httpResp.Body).Decode(provisioner); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal provisioner %q: %v", state.ID.ValueString(), err),
		)
		return
	}

	actual, diags := fromAPI(provisioner, state.AuthorityID.ValueString())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	// Use the original key if it's a jwk provisioner. Terraform will fail if
	// the stored state does not match the current state but there are many valid
	// json serializations of the key.
	if state.JWK != nil && actual.JWK != nil && jsonEqual(state.JWK.Key.ValueString(), actual.JWK.Key.ValueString()) {
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
	httpResp, err := a.client.DeleteProvisioner(ctx, state.AuthorityID.ValueString(), nameOrID, &v20230301.DeleteProvisionerParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete provisioner %s: %v", state.ID.String(), err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Received status %d deleting provisioner %s: %s", httpResp.StatusCode, state.ID.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
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

func jsonEqual(a, b string) bool {
	if a == b {
		return true
	}
	var aVal, bVal any
	if err := json.Unmarshal([]byte(a), &aVal); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bVal); err != nil {
		return false
	}
	return reflect.DeepEqual(aVal, bVal)
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
