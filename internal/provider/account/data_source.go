package account

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

	certInfo, certInfoProps, err := utils.Describe("endpointCertificateInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Certificate Schema",
			err.Error(),
		)
		return
	}

	policy, policyProps, err := utils.Describe("policyMatchCriteria")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Policy Schema",
			err.Error(),
		)
		return
	}

	x509, _, err := utils.Describe("x509Fields")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI X509 Certificate Schema",
			err.Error(),
		)
		return
	}

	ssh, _, err := utils.Describe("sshFields")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI SSH Certificate Schema",
			err.Error(),
		)
		return
	}

	name := schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.StringAttribute{
				Computed: true,
			},
			"device_metadata": schema.StringAttribute{
				Computed: true,
			},
		},
	}

	nameList := schema.SingleNestedAttribute{
		Computed: true,
		Attributes: map[string]schema.Attribute{
			"static": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"device_metadata": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}

	keyInfo, keyInfoProps, err := utils.Describe("endpointKeyInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Key Info Schema",
			err.Error(),
		)
		return
	}

	reloadInfo, reloadInfoProps, err := utils.Describe("endpointReloadInfo")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Reload Info Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: account,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
			},
			"certificate": schema.SingleNestedAttribute{
				MarkdownDescription: certInfo,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"x509": schema.SingleNestedAttribute{
						MarkdownDescription: x509,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"common_name":         name,
							"sans":                nameList,
							"organization":        nameList,
							"organizational_unit": nameList,
							"locality":            nameList,
							"country":             nameList,
							"province":            nameList,
							"street_address":      nameList,
							"postal_code":         nameList,
						},
					},
					"ssh": schema.SingleNestedAttribute{
						MarkdownDescription: ssh,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"key_id":     name,
							"principals": nameList,
						},
					},
					"duration": schema.StringAttribute{
						MarkdownDescription: certInfoProps["duration"],
						Computed:            true,
					},
					"authority_id": schema.StringAttribute{
						MarkdownDescription: certInfoProps["authorityID"],
						Computed:            true,
					},
					"crt_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["crtFile"],
						Computed:            true,
					},
					"key_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["keyFile"],
						Computed:            true,
					},
					"root_file": schema.StringAttribute{
						MarkdownDescription: certInfoProps["rootFile"],
						Computed:            true,
					},
					"uid": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["uid"],
						Computed:            true,
					},
					"gid": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["gid"],
						Computed:            true,
					},
					"mode": schema.Int64Attribute{
						MarkdownDescription: certInfoProps["mode"],
						Computed:            true,
					},
				},
			},
			"key": schema.SingleNestedAttribute{
				MarkdownDescription: keyInfo,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyInfoProps["type"],
					},
					"format": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyInfoProps["format"],
					},
					"pub_file": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyInfoProps["pubFile"],
					},
					"protection": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: keyInfoProps["protection"],
					},
				},
			},
			"reload": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: reloadInfo,
				Attributes: map[string]schema.Attribute{
					"method": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: reloadInfoProps["method"],
					},
					"pid_file": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: reloadInfoProps["pidFile"],
					},
					"signal": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: reloadInfoProps["signal"],
					},
					"unit_name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: reloadInfoProps["unitName"],
					},
				},
			},
			"policy": schema.SingleNestedAttribute{
				Description: policy,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"assurance": schema.ListAttribute{
						MarkdownDescription: policyProps["assurance"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"ownership": schema.ListAttribute{
						MarkdownDescription: policyProps["ownership"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"os": schema.ListAttribute{
						MarkdownDescription: policyProps["operatingSystem"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"source": schema.ListAttribute{
						MarkdownDescription: policyProps["source"],
						ElementType:         types.StringType,
						Computed:            true,
					},
					"tags": schema.ListAttribute{
						MarkdownDescription: policyProps["tags"],
						ElementType:         types.StringType,
						Computed:            true,
					},
				},
			},
			"wifi": schema.SingleNestedAttribute{
				MarkdownDescription: wifi,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: wifiProps["autojoin"],
						Computed:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: wifiProps["caChain"],
						Computed:            true,
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: wifiProps["externalRadiusServer"],
						Computed:            true,
					},
					"hidden": schema.BoolAttribute{
						MarkdownDescription: wifiProps["hidden"],
						Computed:            true,
					},
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: wifiProps["networkAccessServerIP"],
						Computed:            true,
					},
					"ssid": schema.StringAttribute{
						MarkdownDescription: wifiProps["ssid"],
						Computed:            true,
					},
				},
			},
			"vpn": schema.SingleNestedAttribute{
				MarkdownDescription: vpn,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: vpnProps["autojoin"],
						Computed:            true,
					},
					"connection_type": schema.StringAttribute{
						MarkdownDescription: vpnProps["connectionType"],
						Computed:            true,
					},
					"remote_address": schema.StringAttribute{
						MarkdownDescription: vpnProps["remoteAddress"],
						Computed:            true,
					},
					"vendor": schema.StringAttribute{
						MarkdownDescription: vpnProps["vendor"],
						Computed:            true,
					},
					"ike": schema.SingleNestedAttribute{
						MarkdownDescription: ike,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"ca_chain": schema.StringAttribute{
								MarkdownDescription: ikeProps["caChain"],
								Computed:            true,
							},
							"eap": schema.BoolAttribute{
								MarkdownDescription: ikeProps["eap"],
								Computed:            true,
							},
							"remote_id": schema.StringAttribute{
								MarkdownDescription: ikeProps["remoteID"],
								Computed:            true,
							},
						},
					},
				},
			},
			"ethernet": schema.SingleNestedAttribute{
				MarkdownDescription: ethernet,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"autojoin": schema.BoolAttribute{
						MarkdownDescription: ethernetProps["autojoin"],
						Computed:            true,
					},
					"ca_chain": schema.StringAttribute{
						MarkdownDescription: ethernetProps["caChain"],
						Computed:            true,
					},
					"external_radius_server": schema.BoolAttribute{
						MarkdownDescription: ethernetProps["externalRadiusServer"],
						Computed:            true,
					},
					"network_access_server_ip": schema.StringAttribute{
						MarkdownDescription: ethernetProps["networkAccessServerIP"],
						Computed:            true,
					},
				},
			},
			"browser": schema.SingleNestedAttribute{
				MarkdownDescription: browser,
				Computed:            true,
				Attributes:          map[string]schema.Attribute{},
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

	httpResp, err := ds.client.GetAccount(ctx, id, &v20250101.GetAccountParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read account %q: %v", id, err),
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
			fmt.Sprintf("Request %q received status %d reading account %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account := &v20250101.Account{}
	if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal account %s: %v", id, err),
		)
		return
	}

	remote, d := accountFromAPI(ctx, account, req.Config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}
