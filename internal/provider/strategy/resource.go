package strategy

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
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

const name = "smallstep_strategy"

var _ resource.ResourceWithImportState = (*Resource)(nil)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *v20250101.Client
}

func (r *Resource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	strategy, props, err := utils.Describe("protectionStrategy")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Strategy Schema",
			err.Error(),
		)
		return
	}

	cred, credProps, err := utils.Describe("credentialConfigurationRequest")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Credential Strategy Schema",
			err.Error(),
		)
		return
	}

	pol, polProps, err := utils.Describe("policyMatchCriteria")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Policy Strategy Schema",
			err.Error(),
		)
		return
	}

	browser, browserProps, err := utils.Describe("strategyBrowserMutualTLS")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Browser Strategy Schema",
			err.Error(),
		)
		return
	}

	lan, lanProps, err := utils.Describe("strategyLAN")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI LAN Strategy Schema",
			err.Error(),
		)
		return
	}

	relay, relayProps, err := utils.Describe("strategyNetworkRelay")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Network Relay Strategy Schema",
			err.Error(),
		)
		return
	}

	ssh, _, err := utils.Describe("strategySSH")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI SSH Strategy Schema",
			err.Error(),
		)
		return
	}

	sso, ssoProps, err := utils.Describe("strategySSO")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI SSO Strategy Schema",
			err.Error(),
		)
		return
	}

	vpn, vpnProps, err := utils.Describe("strategyVPN")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI VPN Strategy Schema",
			err.Error(),
		)
		return
	}

	ike, ikeProps, err := utils.Describe("ikeV2Config")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI VPN IKEv2 Strategy Schema",
			err.Error(),
		)
		return
	}

	wlan, wlanProps, err := utils.Describe("strategyWLAN")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI WLAN Strategy Schema",
			err.Error(),
		)
		return
	}

	// The objects can be left empty and will be populated with default values.
	// The plan modifier UseStateForUnknown prevents showing (known after apply)
	// for these.
	resp.Schema = schema.Schema{
		MarkdownDescription: strategy,
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
				Required:            true,
			},
			"credential": schema.SingleNestedAttribute{
				MarkdownDescription: cred,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"certificate_info": schema.ObjectAttribute{
						MarkdownDescription: credProps["certificate_info"],
						Optional:            true,
					},
					"key_info": schema.ObjectAttribute{
						MarkdownDescription: credProps["key_info"],
						Optional:            true,
					},
				},
			},
			"policy": schema.SingleNestedAttribute{
				MarkdownDescription: pol,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"assurance": schema.ListAttribute{
						MarkdownDescription: polProps["assurance"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"os": schema.ListAttribute{
						MarkdownDescription: polProps["os"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"ownership": schema.ListAttribute{
						MarkdownDescription: polProps["ownership"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"source": schema.ListAttribute{
						MarkdownDescription: polProps["source"],
						ElementType:         types.StringType,
						Optional:            true,
					},
					"tags": schema.ListAttribute{
						MarkdownDescription: polProps["tags"],
						ElementType:         types.StringType,
						Optional:            true,
					},
				},
			},
			"browser": schema.SingleNestedAttribute{
				MarkdownDescription: browser,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"match_addresses": schema.ListAttribute{
						MarkdownDescription: browserProps["match_addresses"],
						ElementType:         types.StringType,
						Required:            true,
					},
				},
			},
			"ethernet": schema.SingleNestedAttribute{
				MarkdownDescription: lan,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: lanProps["networkAccessServerIP"],
						Optional:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: lanProps["caChain"],
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: lanProps["autojoin"],
						Optional:            true,
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: lanProps["externalRadiusServer"],
						Optional:            true,
					},
				},
			},
			"relay": schema.SingleNestedAttribute{
				MarkdownDescription: relay,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"match_domains": schema.ListAttribute{
						MarkdownDescription: relayProps["match_domains"],
						ElementType:         types.StringType,
						Required:            true,
					},
					"regions": schema.ListAttribute{
						MarkdownDescription: relayProps["regions"],
						ElementType:         types.StringType,
						Required:            true,
					},
				},
			},
			"ssh": schema.SingleNestedAttribute{
				MarkdownDescription: ssh,
				Optional:            true,
				Attributes:          map[string]schema.Attribute{},
			},
			"sso": schema.SingleNestedAttribute{
				MarkdownDescription: sso,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"trusted_roots": schema.StringAttribute{
						MarkdownDescription: ssoProps["trusted_roots"],
						Required:            true,
					},
					"redirect_uri": schema.StringAttribute{
						MarkdownDescription: ssoProps["redirect_uri"],
						Required:            true,
					},
				},
			},
			"vpn": schema.SingleNestedAttribute{
				MarkdownDescription: vpn,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"connection_type": schema.StringAttribute{
						MarkdownDescription: vpnProps["connectionType"],
						Required:            true,
					},
					"vendor": schema.StringAttribute{
						MarkdownDescription: vpnProps["vendor"],
						Optional:            true,
					},
					"remote_address": schema.StringAttribute{
						MarkdownDescription: vpnProps["remoteAddress"],
						Required:            true,
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
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: vpnProps["autojoin"],
						Optional:            true,
					},
				},
			},
			"wifi": schema.SingleNestedAttribute{
				MarkdownDescription: wlan,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: wlanProps["networkAccessServerIP"],
						Optional:            true,
					},
					"ssid": schema.StringAttribute{
						MarkdownDescription: wlanProps["ssid"],
						Required:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: wlanProps["caChain"],
						Computed:            true,
						Optional:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"hidden": schema.BoolAttribute{
						MarkdownDescription: wlanProps["hidden"],
						Optional:            true,
					},
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: wlanProps["autojoin"],
						Optional:            true,
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: wlanProps["externalRadiusServer"],
						Optional:            true,
					},
				},
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

	client, ok := req.ProviderData.(*v20250101.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Get Smallstep API client from provider",
			fmt.Sprintf("Expected *v20250101.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	state := &StrategyModel{}

	resp.Diagnostics.Append(req.State.Get(ctx, state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	strategyID := state.ID.ValueString()
	if strategyID == "" {
		resp.Diagnostics.AddError(
			"Invalid Read Strategy Request",
			"Strategy ID is required.",
		)
		return
	}

	httpResp, err := r.client.GetStrategy(ctx, strategyID, &v20250101.GetStrategyParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read strategy %q: %v", strategyID, err),
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
			fmt.Sprintf("Request %q received status %d reading strategy %s: %s", reqID, httpResp.StatusCode, strategyID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	strategy := &v20250101.ProtectionStrategy{}
	if err := json.NewDecoder(httpResp.Body).Decode(strategy); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal strategy %s: %v", strategyID, err),
		)
		return
	}

	remote, d := fromAPI(ctx, strategy, req.State)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	plan := &StrategyModel{}
	resp.Diagnostics.Append(req.Plan.Get(ctx, plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	strategy, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	reqBody := v20250101.ProtectionStrategyRequest{
		Configuration: v20250101.ProtectionStrategyRequest_Configuration(strategy.Configuration),
		Credential:    strategy.Credential,
		Kind:          strategy.Kind,
		Name:          strategy.Name,
		Policy:        strategy.Policy,
	}

	httpResp, err := r.client.PostStrategies(ctx, &v20250101.PostStrategiesParams{}, reqBody)
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to create strategy: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusCreated {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d creating strategy: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	strategy = &v20250101.ProtectionStrategy{}
	if err := json.NewDecoder(httpResp.Body).Decode(strategy); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal strategy: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, strategy, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan := &StrategyModel{}
	diags := req.Plan.Get(ctx, plan)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	strategyID := plan.ID.ValueString()

	strategy, diags := toAPI(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	httpResp, err := r.client.PutStrategy(ctx, strategyID, &v20250101.PutStrategyParams{}, *strategy)
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
			fmt.Sprintf("Request %q received status %d updating strategy: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	strategy = &v20250101.ProtectionStrategy{}
	if err := json.NewDecoder(httpResp.Body).Decode(strategy); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to parse strategy update response: %v", err),
		)
		return
	}

	model, diags := fromAPI(ctx, strategy, req.Plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	state := &StrategyModel{}
	diags := req.State.Get(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	strategyID := state.ID.ValueString()
	if strategyID == "" {
		resp.Diagnostics.AddError(
			"Invalid Delete Strategy Request",
			"Strategy ID is required",
		)
		return
	}

	httpResp, err := r.client.DeleteStrategy(ctx, strategyID, &v20250101.DeleteStrategyParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to delete strategy: %v", err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d deleting strategy: %s", reqID, httpResp.StatusCode, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}
}

func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
