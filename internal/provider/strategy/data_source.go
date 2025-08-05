package strategy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	v20250101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20250101"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/certinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/models/certificates/keyinfo"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

type DataSource struct {
	client *v20250101.Client
}

func (ds *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = name
}

// Configure adds the Smallstep API client to the data source.
func (ds *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	ds.client = client
}

func (ds *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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

	relay, relayProps, err := utils.Describe("strategyNetworkRelayConfig")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Network Relay Strategy Schema",
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

	certInfo, err := certinfo.NewDataSourceSchema()
	if err != nil {
		resp.Diagnostics.AddError(
			"Certificate Info",
			err.Error(),
		)
		return
	}

	keyInfo, err := keyinfo.NewDataSourceSchema()
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
						Computed:            true,
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
						MarkdownDescription: relayProps["matchDomains"],
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
						MarkdownDescription: ssoProps["trustedRoots"],
						Required:            true,
					},
					"redirect_uri": schema.StringAttribute{
						MarkdownDescription: ssoProps["redirectUri"],
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

func (ds *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id string
	diags := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := ds.client.GetProtectionStrategy(ctx, id, &v20250101.GetProtectionStrategyParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read strategy %q: %v", id, err),
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
			fmt.Sprintf("Request %q received status %d reading strategy %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	strategy := &v20250101.ProtectionStrategy{}
	if err := json.NewDecoder(httpResp.Body).Decode(strategy); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal strategy %s: %v", id, err),
		)
		return
	}

	remote, d := fromAPI(ctx, strategy, req.Config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}
