package collection_instance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	v20231101 "github.com/smallstep/terraform-provider-smallstep/internal/apiclient/v20231101"
	"github.com/smallstep/terraform-provider-smallstep/internal/provider/utils"
)

var _ datasource.DataSourceWithConfigure = (*DataSource)(nil)

func NewDataSource() datasource.DataSource {
	return &DataSource{}
}

// DataSource implements data.smallstep_collection
type DataSource struct {
	client *v20231101.Client
}

func (a *DataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = instanceTypeName
}

// Configure adds the Smallstep API client to the data source.
func (a *DataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	a.client = client
}

func (a *DataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config Model

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	slug := config.CollectionSlug.ValueString()
	id := config.ID.ValueString()

	httpResp, err := a.client.GetCollectionInstance(ctx, slug, id, &v20231101.GetCollectionInstanceParams{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to read collection instance %s/%s: %v", slug, id, err),
		)
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		reqID := httpResp.Header.Get("X-Request-Id")
		resp.Diagnostics.AddError(
			"Smallstep API Response Error",
			fmt.Sprintf("Request %q received status %d reading collection instance %s/%s: %s", reqID, httpResp.StatusCode, slug, id, utils.APIErrorMsg(httpResp.Body)),
		)
		return
	}

	instance := &v20231101.CollectionInstance{}
	if err := json.NewDecoder(httpResp.Body).Decode(instance); err != nil {
		resp.Diagnostics.AddError(
			"Smallstep API Client Error",
			fmt.Sprintf("Failed to unmarshal collection instance %s/%s: %v", slug, id, err),
		)
		return
	}

	remote, d := fromAPI(ctx, slug, instance, req.Config)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("read collection instance %s/%s data source", slug, id))

	resp.Diagnostics.Append(resp.State.Set(ctx, &remote)...)
}

func (d *DataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	component, props, err := utils.Describe("collectionInstance")
	if err != nil {
		resp.Diagnostics.AddError(
			"Parse Smallstep OpenAPI spec",
			err.Error(),
		)
		return
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: component,

		Attributes: map[string]schema.Attribute{
			"collection_slug": schema.StringAttribute{
				MarkdownDescription: "The slug of the collection this instance belongs to",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: props["id"],
				Required:            true,
			},
			"data": schema.StringAttribute{
				MarkdownDescription: props["data"],
				Computed:            true,
			},
			"out_data": schema.StringAttribute{
				MarkdownDescription: props["data"],
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: props["createdAt"],
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: props["updatedAt"],
				Computed:            true,
			},
		},
	}
}
