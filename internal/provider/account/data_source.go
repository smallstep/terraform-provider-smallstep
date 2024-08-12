package account

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSource = &DataSource{}

type DataSource struct {
	client *v20231101.Client
}

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

func (d *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = accountTypeName
}

// Configure adds the Smallstep API client to the data source.
func (d *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	d.client = client
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
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
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						utils.UUID,
						"must be a valid UUID",
					),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
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

func (d *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var id string

	ds := req.Config.GetAttribute(ctx, path.Root("id"), &id)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetAccount(ctx, id, &v20231101.GetAccountParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read account %q: %v", id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	switch httpResp.StatusCode {
	case http.StatusOK:
		break
	case http.StatusNotFound:
		resp.Diagnostics.AddError(
			"Account Not Found",
			fmt.Sprintf("Account %q data source not found", id),
		)
		return
	default:
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading account %q: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	account := &v20231101.Account{}
	if err := json.NewDecoder(httpResp.Body).Decode(account); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal account %s: %v", id, err),
		)
		return
	}

	remote, ds := fromAPI(ctx, account, req.Config)
	resp.Diagnostics.Append(ds...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}
