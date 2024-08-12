package account

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
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

// Resource defines the resource implementation.
type Resource struct {
	client *v20231101.Client
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	account, props, err := utils.Describe("account")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Account Schema",
			err.Error(),
		)
		return
	}

	wifi, wifiProps, err := utils.Describe("wifiAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI WiFi Account Schema",
			err.Error(),
		)
		return
	}

	vpn, vpnProps, err := utils.Describe("vpnAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI VPN Account Schema",
			err.Error(),
		)
		return
	}

	ike, ikeProps, err := utils.Describe("ikeV2Config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI VPN IKEv2 Account Schema",
			err.Error(),
		)
		return
	}

	browser, _, err := utils.Describe("browserAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Browser Account Schema",
			err.Error(),
		)
		return
	}

	ethernet, ethernetProps, err := utils.Describe("ethernetAccount")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Ethernet Account Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: account,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Required:            true,
			},
			"wifi": schema.SingleNestedAttribute{
				MarkdownDescription: wifi,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: wifiProps["autojoin"],
						Optional:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: wifiProps["caChain"],
						Computed:            true,
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: wifiProps["externalRadiusServer"],
						Optional:            true,
					},
					"hidden": schema.BoolAttribute{
						MarkdownDescription: wifiProps["hidden"],
						Optional:            true,
					},
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: wifiProps["networkAccessServerIP"],
						Optional:            true,
					},
					"ssid": schema.StringAttribute{
						MarkdownDescription: wifiProps["ssid"],
						Required:            true,
					},
				},
			},
			"vpn": schema.SingleNestedAttribute{
				MarkdownDescription: vpn,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: vpnProps["autojoin"],
						Optional:            true,
					},
					"connection_type": schema.StringAttribute{
						MarkdownDescription: vpnProps["connectionType"],
						Required:            true,
					},
					"remote_address": schema.StringAttribute{
						MarkdownDescription: vpnProps["remoteAddress"],
						Required:            true,
					},
					"vendor": schema.StringAttribute{
						MarkdownDescription: vpnProps["vendor"],
						Optional:            true,
					},
					"ike": schema.SingleNestedAttribute{
						MarkdownDescription: ike,
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"ca_chain": schema.StringAttribute{
								MarkdownDescription: ikeProps["caChain"],
								Optional:            true,
							},
							"eap": schema.BoolAttribute{
								MarkdownDescription: ikeProps["eap"],
								Optional:            true,
							},
							"remote_id": schema.StringAttribute{
								MarkdownDescription: ikeProps["remoteID"],
								Optional:            true,
							},
						},
					},
				},
			},
			"ethernet": schema.SingleNestedAttribute{
				MarkdownDescription: ethernet,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: ethernetProps["autojoin"],
						Optional:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: ethernetProps["caChain"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: ethernetProps["externalRadiusServer"],
						Optional:            true,
					},
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: ethernetProps["networkAccessServerIP"],
						Optional:            true,
					},
				},
			},
			"browser": schema.SingleNestedAttribute{
				MarkdownDescription: browser,
				Optional:            true,
				Attributes:          map[string]schema.Attribute{},
			},
		},
	}
}

func (r *Resource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = accountTypeName
}

// Configure adds the Smallstep API client to the resource.
func (r *Resource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*v20231101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20231101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &Model{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := r.client.GetAccount(ctx, state.ID.ValueString(), &v20231101.GetAccountParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read account %q: %v", state.ID.ValueString(), err),
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
			fmt.Sprintf("Request %q received status %d reading account %s: %s", reqID, httpResp.StatusCode, state.ID.String(), utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account := &v20231101.Account{}
	if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal account %s: %v", state.ID.String(), err),
		)
		return
	}

	remote, d := fromAPI(ctx, account, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (a *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &Model{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqBody, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := a.client.PostAccounts(ctx, &v20231101.PostAccountsParams{}, *reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create account: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating account: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account := &v20231101.Account{}
	if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal account: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, account, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &Model{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	var accountID string
	req.State.GetAttribute(ctx, path.Root("id"), &accountID)

	account, diags := toAPI(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	account.Id = &accountID

	httpResp, err := r.client.PutAccount(ctx, accountID, &v20231101.PutAccountParams{}, *account)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			err.Error(),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating account: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account = &v20231101.Account{}
	if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse account update response: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, account, req.Plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &Model{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	accountID := state.ID.ValueString()
	if accountID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Account Request",
			"Account ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteAccount(ctx, accountID, &v20231101.DeleteAccountParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete account: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d updating account: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
