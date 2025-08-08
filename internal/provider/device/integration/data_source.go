package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"

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
	integration, props, err := utils.Describe("deviceInventoryIntegration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Inventory Integration Schema",
			err.Error(),
		)
		return
	}

	jamf, jamfProps, err := utils.Describe("jamfInventoryIntegration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Jamf Device Inventory Integration Schema",
			err.Error(),
		)
		return
	}

	intune, intuneProps, err := utils.Describe("intuneInventoryIntegration")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Intune Device Inventory Integration Schema",
			err.Error(),
		)
		return
	}

	// The objects can be left empty and will be populated with default values.
	// The plan modifier UseStateForUnknown prevents showing (known after apply)
	// for these.
	resp.Schema = schema.Schema{
		MarkdownDescription: integration,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Optional:            true,
			},
			"jamf": schema.SingleNestedAttribute{
				MarkdownDescription: jamf,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"client_id": schema.StringAttribute{
						MarkdownDescription: jamfProps["clientId"],
						Optional:            true,
					},
					"client_secret": schema.StringAttribute{
						MarkdownDescription: jamfProps["clientSecret"],
						Optional:            true,
					},
					"tenant_url": schema.StringAttribute{
						MarkdownDescription: jamfProps["tenantUrl"],
						Required:            true,
					},
				},
			},
			"intune": schema.SingleNestedAttribute{
				MarkdownDescription: intune,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"app_id": schema.StringAttribute{
						MarkdownDescription: intuneProps["appId"],
						Required:            true,
					},
					"app_secret": schema.StringAttribute{
						MarkdownDescription: intuneProps["appSecret"],
						Required:            true,
					},
					"azure_tenant_name": schema.StringAttribute{
						MarkdownDescription: intuneProps["azureTenantName"],
						Required:            true,
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

	httpResp, err := ds.client.GetDeviceInventoryIntegration(ctx, id, &v20250101.GetDeviceInventoryIntegrationParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read device inventory integration %q: %v", id, err),
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
			fmt.Sprintf("Request %q received status %d reading device inventory integration %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	inventory := &v20250101.DeviceInventoryIntegration{}
	if err := json.NewDecoder(httpResp.Body).Decode(inventory); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device inventory integration %s: %v", id, err),
		)
		return
	}

	remote, d := fromAPI(ctx, inventory, req.Config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}
