package vpn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/apiclient/clientset"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *v20250101.Client
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	vpn, props, err := utils.Describe("vpn")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI VPN Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: vpn,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"connection_type": schema.StringAttribute{
				MarkdownDescription: props["connectionType"],
				Required:            true,
			},
			"remote_address": schema.StringAttribute{
				MarkdownDescription: props["remoteAddress"],
				Required:            true,
			},
			"autojoin": schema.BoolAttribute{
				MarkdownDescription: props["autojoin"],
				Optional:            true,
			},
			"vendor": schema.StringAttribute{
				MarkdownDescription: props["vendor"],
				Optional:            true,
			},
			"ike": schema.SingleNestedAttribute{
				MarkdownDescription: props["ike"],
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: "The certificate authority bundle that client certificates must chain up to.",
						Required:            true,
					},
					"eap": schema.BoolAttribute{
						MarkdownDescription: "Whether or not EAP is enforced on this VPN server.",
						Optional:            true,
					},
					"remote_id": schema.StringAttribute{
						MarkdownDescription: "Typically, the common name of the remote server. Defaults to the remote address.",
						Optional:            true,
					},
				},
			},
			"credentials": schema.SetAttribute{
				MarkdownDescription: props["credentials"],
				ElementType:         types.StringType,
				Optional:            true,
			},
		},
	}
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = name
}

// Configure adds the Smallstep API client to the resource.
func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*clientset.Clients)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *clientset.Clients, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = clients.V20250101
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &VpnModel{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpnID := state.ID.ValueString()
	if vpnID == "" {
		resp.Diagnostics.AddError(
			"Invalid Read VPN Request",
			"VPN ID is required.",
		)
		return
	}

	httpResp, err := r.client.GetVpn(ctx, vpnID, &v20250101.GetVpnParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read vpn %q: %v", vpnID, err),
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
			fmt.Sprintf("Request %q received status %d reading vpn %s: %s", reqID, httpResp.StatusCode, vpnID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	vpn := &v20250101.Vpn{}
	if err := json.NewDecoder(httpResp.Body).Decode(vpn); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal vpn %s: %v", vpnID, err),
		)
		return
	}

	remote := FromAPI(ctx, vpn, &resp.Diagnostics, req.State)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &VpnModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody := plan.ToAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PostVpn(ctx, &v20250101.PostVpnParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create vpn: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating vpn: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	vpn := &v20250101.Vpn{}
	if err := json.NewDecoder(httpResp.Body).Decode(vpn); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal vpn: %v", err),
		)
		return
	}

	model := FromAPI(ctx, vpn, &resp.Diagnostics, req.Plan)
	if resp.Diagnostics.HasError() {
		return
	}
	diags := resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &VpnModel{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	vpnID := plan.ID.ValueString()

	reqBody := plan.ToAPI(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.PutVpn(ctx, vpnID, &v20250101.PutVpnParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			err.Error(),
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
			fmt.Sprintf("Request %q received status %d updating vpn: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	vpn := &v20250101.Vpn{}
	if err := json.NewDecoder(httpResp.Body).Decode(vpn); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse vpn update response: %v", err),
		)
		return
	}

	model := FromAPI(ctx, vpn, &resp.Diagnostics, req.Plan)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &VpnModel{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vpnID := state.ID.ValueString()
	if vpnID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete VPN Request",
			"VPN ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteVpn(ctx, vpnID, &v20250101.DeleteVpnParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete vpn: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting vpn: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
