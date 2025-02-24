package device

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
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

func (a *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = typeName
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
	device, props, err := utils.Describe("device")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device Schema",
			err.Error(),
		)
		return
	}

	deviceUser, userProps, err := utils.Describe("deviceUser")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Device User Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: device,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"permanent_identifier": schema.StringAttribute{
				MarkdownDescription: props["permanentIdentifier"],
				Computed:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: props["serial"],
				Computed:            true,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: props["displayName"],
				Computed:            true,
			},
			"display_id": schema.StringAttribute{
				MarkdownDescription: props["displayId"],
				Computed:            true,
			},
			"os": schema.StringAttribute{
				MarkdownDescription: props["os"],
				Computed:            true,
			},
			"ownership": schema.StringAttribute{
				MarkdownDescription: props["ownership"],
				Computed:            true,
			},
			"metadata": schema.MapAttribute{
				MarkdownDescription: props["metadata"],
				Computed:            true,
				ElementType:         types.StringType,
			},
			"tags": schema.SetAttribute{
				MarkdownDescription: props["tags"],
				Computed:            true,
				ElementType:         types.StringType,
			},
			"user": schema.SingleNestedAttribute{
				MarkdownDescription: deviceUser,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"display_name": schema.StringAttribute{
						MarkdownDescription: userProps["displayName"],
						Computed:            true,
					},
					"email": schema.StringAttribute{
						MarkdownDescription: userProps["email"],
						Computed:            true,
					},
				},
			},
			"connected": schema.BoolAttribute{
				MarkdownDescription: props["connected"],
				Computed:            true,
			},
			"high_assurance": schema.BoolAttribute{
				MarkdownDescription: props["highAssurance"],
				Computed:            true,
			},
			"enrolled_at": schema.StringAttribute{
				MarkdownDescription: props["enrolledAt"],
				Computed:            true,
			},
			"approved_at": schema.StringAttribute{
				MarkdownDescription: props["approvedAt"],
				Computed:            true,
			},
			"last_seen": schema.StringAttribute{
				MarkdownDescription: props["lastSeen"],
				Computed:            true,
			},
		},
	}
}

func (ds *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Model

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	deviceID := config.ID.ValueString()
	if deviceID == "" {
		resp.Diagnostics.AddError(
			"Invalid Device ID",
			"Device ID is required",
		)
		return
	}

	httpResp, err := ds.client.GetDevice(ctx, deviceID, &v20250101.GetDeviceParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read device %q: %v", config.ID.ValueString(), err),
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
			fmt.Sprintf("Request %q received status %d reading device %s: %s", reqID, httpResp.StatusCode, deviceID, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	device := &v20250101.Device{}
	if err := json.NewDecoder(httpResp.Body).Decode(device); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal device %s: %v", deviceID, err),
		)
		return
	}

	remote, d := fromAPI(ctx, device, req.Config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}
