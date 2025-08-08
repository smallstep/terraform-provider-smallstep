package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/certinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/keyinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/x509info"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/strategies/relay"
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

	cred, _, err := utils.Describe("credentialConfigurationRequest")
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

	browser, browserProps, err := utils.Describe("strategyBrowserMutualTLSConfig")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Browser Strategy Schema",
			err.Error(),
		)
		return
	}

	lan, lanProps, err := utils.Describe("strategyLANConfig")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI LAN Strategy Schema",
			err.Error(),
		)
		return
	}

	relayDesc, relayProps, err := utils.Describe("strategyNetworkRelayConfig")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Network Relay Strategy Schema",
			err.Error(),
		)
		return
	}

	_, relayServerProps, err := utils.Describe("strategyNetworkRelayServer")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Network Relay Server Strategy Schema",
			err.Error(),
		)
		return
	}

	ssh, _, err := utils.Describe("strategySSHConfig")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI SSH Strategy Schema",
			err.Error(),
		)
		return
	}

	sso, ssoProps, err := utils.Describe("strategySSOConfig")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI SSO Strategy Schema",
			err.Error(),
		)
		return
	}

	_, ssoClientProps, err := utils.Describe("deviceIdentityProviderClient")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Network SSO Client Strategy Schema",
			err.Error(),
		)
		return
	}

	_, ssoIdentityProviderProps, err := utils.Describe("deviceIdentityProvider")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Network SSO Identity Provider Strategy Schema",
			err.Error(),
		)
		return
	}

	vpn, vpnProps, err := utils.Describe("strategyVPNConfig")
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

	wlan, wlanProps, err := utils.Describe("strategyWLANConfig")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI WLAN Strategy Schema",
			err.Error(),
		)
		return
	}

	_, networkProps, err := utils.Describe("wirelessNetwork")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI WLAN Network Strategy Schema",
			err.Error(),
		)
		return
	}

	_, radiusProps, err := utils.Describe("radiusServer")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI WLAN Radius Strategy Schema",
			err.Error(),
		)
		return
	}

	certInfo, err := certinfo.NewResourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Certificate Info",
			err.Error(),
		)
		return
	}

	keyInfo, err := keyinfo.NewResourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Key Info",
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
					"certificate_info": certInfo,
					"key_info":         keyInfo,
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
						MarkdownDescription: browserProps["matchAddresses"],
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
					},
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: lanProps["autojoin"],
						Optional:            true,
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: lanProps["externalRadiusServer"],
						Optional:            true,
					},
					"radius": schema.SingleNestedAttribute{
						MarkdownDescription: wlanProps["radius"],
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"ca_chain": schema.StringAttribute{
								MarkdownDescription: radiusProps["ca_chain"],
								Computed:            true,
							},
							"ip_addresses": schema.ListAttribute{
								MarkdownDescription: radiusProps["ip_addresses"],
								Computed:            true,
								ElementType:         types.StringType,
							},
						},
					},
				},
			},
			"relay": schema.SingleNestedAttribute{
				MarkdownDescription: relayDesc,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"match_domains": schema.ListAttribute{
						MarkdownDescription: relayProps["matchDomains"],
						ElementType:         types.StringType,
						Required:            true,
					},
					"regions": schema.ListAttribute{
						MarkdownDescription: relayProps["regions"],
						ElementType:         types.StringType,
						Required:            true,
					},
					"proxy_instances": schema.ListAttribute{
						MarkdownDescription: relayProps["proxy_instances"],
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						ElementType: types.ObjectType{AttrTypes: relay.ProxyInstanceAttributes},
						Computed:    true,
					},
					"server": schema.SingleNestedAttribute{
						MarkdownDescription: relayProps["server"],
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"ca_chain": schema.StringAttribute{
								MarkdownDescription: relayServerProps["ca_chain"],
								Computed:            true,
							},
							"hostname": schema.StringAttribute{
								MarkdownDescription: relayServerProps["hostname"],
								Computed:            true,
							},
						},
						Computed: true,
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
						MarkdownDescription: ssoProps["trustedRoots"],
						Required:            true,
					},
					"redirect_uri": schema.StringAttribute{
						MarkdownDescription: ssoProps["redirectUri"],
						Required:            true,
					},
					"client": schema.SingleNestedAttribute{
						MarkdownDescription: ssoProps["client"],
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"id": schema.StringAttribute{
								MarkdownDescription: ssoClientProps["id"],
								Computed:            true,
							},
							"redirect_uri": schema.StringAttribute{
								MarkdownDescription: ssoClientProps["redirect_uri"],
								Computed:            true,
							},
							"secret": schema.StringAttribute{
								MarkdownDescription: ssoClientProps["secret"],
								Computed:            true,
								Sensitive:           true,
							},
						},
					},
					"identity_provider": schema.SingleNestedAttribute{
						MarkdownDescription: ssoProps["identity_provider"],
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"authorize_endpoint": schema.StringAttribute{
								MarkdownDescription: ssoIdentityProviderProps["authorize_endpoint"],
								Computed:            true,
							},
							"jwks_endpoint": schema.StringAttribute{
								MarkdownDescription: ssoIdentityProviderProps["jwks_endpoint"],
								Computed:            true,
							},
							"trust_roots": schema.StringAttribute{
								MarkdownDescription: ssoIdentityProviderProps["trust_roots"],
								Computed:            true,
							},
						},
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
						Optional:            true,
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
					"network": schema.SingleNestedAttribute{
						MarkdownDescription: wlanProps["network"],
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"ssid": schema.StringAttribute{
								MarkdownDescription: networkProps["ssid"],
								Computed:            true,
							},
							"hidden": schema.BoolAttribute{
								MarkdownDescription: networkProps["hidden"],
								Computed:            true,
							},
							"autojoin": schema.BoolAttribute{
								MarkdownDescription: networkProps["autojoin"],
								Computed:            true,
							},
						},
					},
					"radius": schema.SingleNestedAttribute{
						MarkdownDescription: wlanProps["radius"],
						Computed:            true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"ca_chain": schema.StringAttribute{
								MarkdownDescription: radiusProps["ca_chain"],
								Computed:            true,
							},
							"ip_addresses": schema.ListAttribute{
								MarkdownDescription: radiusProps["ip_addresses"],
								Computed:            true,
								ElementType:         types.StringType,
							},
						},
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

	debug("Read request")
	debug(strategyID)

	httpResp, err := r.client.GetProtectionStrategy(ctx, strategyID, &v20250101.GetProtectionStrategyParams{})
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

	debug("Read response:")
	debug(strategy)

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

	var credential *v20250101.CredentialConfigurationRequest
	if e := strategy.EndpointConfiguration; e.CertificateInfo != nil || e.KeyInfo != nil {
		credential = &v20250101.CredentialConfigurationRequest{
			CertificateInfo: e.CertificateInfo,
			KeyInfo:         e.KeyInfo,
		}
	}

	reqBody := v20250101.ProtectionStrategyRequest{
		Configuration: v20250101.ProtectionStrategyRequest_Configuration(strategy.Configuration),
		Credential:    credential,
		Kind:          strategy.Kind,
		Name:          strategy.Name,
		Policy:        strategy.EndpointConfiguration.Policy,
	}

	debug("Create request:")
	debug(reqBody)

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

	debug("Create response:")
	debug(strategy)

	model, diags := fromAPI(ctx, strategy, req.Plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)

	// Track whether state holds a computed value. Required for our
	// MaybeUseStateForUnknown object plan modifier.
	x509 := types.Object{}
	req.Config.GetAttribute(ctx, path.Root("credential").AtName("certificate_info").AtName("x509"), &x509)
	if x509.IsNull() {
		diags = resp.Private.SetKey(ctx, x509info.X509PrivateKey, x509info.Computed)
		resp.Diagnostics.Append(diags...)
	} else {
		diags = resp.Private.SetKey(ctx, x509info.X509PrivateKey, nil)
		resp.Diagnostics.Append(diags...)
	}
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

	var credential *v20250101.CredentialConfigurationRequest
	if e := strategy.EndpointConfiguration; e.CertificateInfo != nil || e.KeyInfo != nil {
		credential = &v20250101.CredentialConfigurationRequest{
			CertificateInfo: e.CertificateInfo,
			KeyInfo:         e.KeyInfo,
		}
	}

	reqBody := v20250101.ProtectionStrategyUpdateRequest{
		Id:            strategyID,
		Credential:    credential,
		Kind:          strategy.Kind,
		Name:          strategy.Name,
		Configuration: v20250101.ProtectionStrategyUpdateRequest_Configuration(strategy.Configuration),
		Policy:        strategy.EndpointConfiguration.Policy,
	}

	debug("Update request:")
	debug(reqBody)

	httpResp, err := r.client.PutStrategy(ctx, strategyID, &v20250101.PutStrategyParams{}, reqBody)
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

	debug("Update response:")
	debug(strategy)

	model, diags := fromAPI(ctx, strategy, req.Plan)
	resp.Diagnostics.Append(diags...)

	diags = resp.State.Set(ctx, model)
	resp.Diagnostics.Append(diags...)

	// Track whether state holds a computed value. Required for our
	// MaybeUseStateForUnknown object plan modifier.
	x509 := types.Object{}
	req.Config.GetAttribute(ctx, path.Root("credential").AtName("certificate_info").AtName("x509"), &x509)
	if x509.IsNull() {
		diags = resp.Private.SetKey(ctx, x509info.X509PrivateKey, x509info.Computed)
		resp.Diagnostics.Append(diags...)
	} else {
		diags = resp.Private.SetKey(ctx, x509info.X509PrivateKey, nil)
		resp.Diagnostics.Append(diags...)
	}
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

	debug("Delete request:")
	debug(strategyID)

	httpResp, err := r.client.DeleteProtectionStrategy(ctx, strategyID, &v20250101.DeleteProtectionStrategyParams{})
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

func debug(v any) {
	switch v := v.(type) {
	case string:
		fmt.Println(v)
	case []byte:
		fmt.Println(string(v))
	default:
		b, _ := json.Marshal(v)
		fmt.Println(string(b))
	}
}
