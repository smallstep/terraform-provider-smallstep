package managed_radius

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

func (a *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
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

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	radius, props, err := utils.Describe("managedRadius")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Managed RADIUS Schema",
			err.Error(),
		)
		return
	}

	replyAttrs, replyAttrsProps, err := utils.Describe("replyAttribute")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI Reply Attributes Schema",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: radius,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: props["name"],
				Computed:            true,
			},
			"nas_ips": schema.ListAttribute{
				MarkdownDescription: props["nasIPs"],
				Computed:            true,
				ElementType:         types.StringType,
			},
			"client_ca": schema.StringAttribute{
				MarkdownDescription: props["clientCA"],
				Computed:            true,
			},
			"reply_attributes": schema.ListNestedAttribute{
				MarkdownDescription: replyAttrs,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: replyAttrsProps["name"],
						},
						"value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: replyAttrsProps["value"],
						},
						"value_from_certificate": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: replyAttrsProps["valueFromCertificate"],
						},
					},
				},
			},

			"server_ca": schema.StringAttribute{
				MarkdownDescription: props["serverCA"],
				Computed:            true,
			},
			"server_ip": schema.StringAttribute{
				MarkdownDescription: props["serverIP"],
				Computed:            true,
			},
			"server_port": schema.StringAttribute{
				MarkdownDescription: props["serverPort"],
				Computed:            true,
			},
			"server_hostname": schema.StringAttribute{
				MarkdownDescription: props["serverHostname"],
				Computed:            true,
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

	httpResp, err := ds.client.GetManagedRadius(ctx, id, &v20250101.GetManagedRadiusParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read managed radius %q: %v", id, err),
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
			fmt.Sprintf("Request %q received status %d reading managed radius %s: %s", reqID, httpResp.StatusCode, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	managedRadius := &v20250101.ManagedRadius{}
	if err := json.NewDecoder(httpResp.Body).Decode(managedRadius); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal managed radius %s: %v", id, err),
		)
		return
	}

	remote := fromAPI(ctx, &resp.Diagnostics, managedRadius, req.Config)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, remote)...)
}
